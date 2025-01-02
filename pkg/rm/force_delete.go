package rm

import (
    "context"
    "fmt"
    "strings"

    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime/schema"
    "k8s.io/apimachinery/pkg/types"
    "k8s.io/client-go/dynamic"
    "k8s.io/client-go/kubernetes"
    "k8s.io/cli-runtime/pkg/genericclioptions"
)

func ForceDelete(flags *genericclioptions.ConfigFlags, streams genericclioptions.IOStreams, args []string) error {
    if len(args) == 0 {
        return fmt.Errorf("resource argument is required")
    }

    // Get REST config from flags
    config, err := flags.ToRESTConfig()
    if err != nil {
        return fmt.Errorf("failed to get REST config: %v", err)
    }

    // Create the clientset
    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        return fmt.Errorf("failed to create kubernetes client: %v", err)
    }

    // Create dynamic client
    dynamicClient, err := dynamic.NewForConfig(config)
    if err != nil {
        return fmt.Errorf("failed to create dynamic client: %v", err)
    }

    return forceDelete(context.Background(), args[0], flags, dynamicClient, clientset)
}

func forceDelete(ctx context.Context, resource string, flags *genericclioptions.ConfigFlags, dynamicClient dynamic.Interface, clientset *kubernetes.Clientset) error {
    parts := strings.Split(resource, "/")
    if len(parts) != 2 {
        return fmt.Errorf("invalid resource format. Expected <type>/<name>, got %s", resource)
    }

    resourceType := parts[0]
    resourceName := parts[1]

    // Get namespace from flags
    namespace := ""
    if ns := flags.Namespace; ns != nil {
        namespace = *ns
    }

    // Handle namespace resources differently
    if resourceType == "namespace" || resourceType == "ns" {
        return forceDeleteNamespace(ctx, resourceName, clientset)
    }

    // Get GVR for the resource type
    gvr, err := getGroupVersionResource(resourceType, clientset)
    if err != nil {
        return err
    }

    fmt.Printf("🔍 Processing %s/%s in namespace %s...\n", resourceType, resourceName, namespace)

    // Remove finalizers
    fmt.Printf("🔧 Removing finalizers...\n")
    patch := []byte(`{"metadata":{"finalizers":null}}`)
    _, err = dynamicClient.Resource(gvr).Namespace(namespace).Patch(ctx, resourceName, 
        types.MergePatchType, patch, metav1.PatchOptions{})
    if err != nil {
        fmt.Printf("⚠️  Warning: Failed to remove finalizers: %v\n", err)
    }

    // Force delete
    fmt.Printf("🗑️  Force deleting resource...\n")
    deleteOptions := metav1.DeleteOptions{
        GracePeriodSeconds: new(int64),
    }
    err = dynamicClient.Resource(gvr).Namespace(namespace).Delete(ctx, resourceName, deleteOptions)
    if err != nil {
        return fmt.Errorf("failed to delete resource: %v", err)
    }

    fmt.Printf("✅ Successfully initiated force deletion of %s/%s\n", resourceType, resourceName)
    return nil
}

func forceDeleteNamespace(ctx context.Context, name string, clientset *kubernetes.Clientset) error {
    fmt.Printf("🔍 Processing namespace %s...\n", name)

    // Get the namespace
    ns, err := clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
    if err != nil {
        return fmt.Errorf("failed to get namespace: %v", err)
    }

    // Check if namespace is in Terminating state
    if ns.Status.Phase != "Terminating" {
        return fmt.Errorf("namespace %s is not in Terminating state (current state: %s)", name, ns.Status.Phase)
    }

    // Check if namespace has a deletion timestamp
    if ns.DeletionTimestamp == nil {
        return fmt.Errorf("namespace %s is not being deleted (no deletion timestamp)", name)
    }

    // Create a finalize request by removing the finalizers
    fmt.Printf("🔧 Removing finalizers via finalize endpoint...\n")
    nsFinalize := ns.DeepCopy()
    nsFinalize.Spec.Finalizers = nil
    
    // Call finalize endpoint
    _, err = clientset.CoreV1().RESTClient().Put().
        Resource("namespaces").
        Name(name).
        SubResource("finalize").
        Body(nsFinalize).
        Do(ctx).
        Raw()
    if err != nil {
        fmt.Printf("⚠️  Warning: Failed to finalize namespace: %v\n", err)
    }

    // Force delete with 0 grace period
    fmt.Printf("🗑️  Force deleting namespace...\n")
    zero := int64(0)
    deleteOptions := metav1.DeleteOptions{
        GracePeriodSeconds: &zero,
        PropagationPolicy: &[]metav1.DeletionPropagation{metav1.DeletePropagationBackground}[0],
    }
    err = clientset.CoreV1().Namespaces().Delete(ctx, name, deleteOptions)
    if err != nil {
        // If the namespace is not found, that's actually good - it means it was successfully removed
        if strings.Contains(err.Error(), "not found") {
            fmt.Printf("✅ Successfully removed namespace %s\n", name)
            return nil
        }
        return fmt.Errorf("failed to delete namespace: %v", err)
    }

    fmt.Printf("✅ Successfully initiated force deletion of namespace %s\n", name)
    return nil
}

func getGroupVersionResource(resourceType string, clientset *kubernetes.Clientset) (schema.GroupVersionResource, error) {
    // Common resource types mapping
    resourceMap := map[string]schema.GroupVersionResource{
        "pod":         {Version: "v1", Resource: "pods"},
        "deployment": {Group: "apps", Version: "v1", Resource: "deployments"},
        "service":    {Version: "v1", Resource: "services"},
        "configmap":  {Version: "v1", Resource: "configmaps"},
        "secret":     {Version: "v1", Resource: "secrets"},
        "pvc":        {Version: "v1", Resource: "persistentvolumeclaims"},
        "pv":         {Version: "v1", Resource: "persistentvolumes"},
    }

    if gvr, ok := resourceMap[strings.ToLower(resourceType)]; ok {
        return gvr, nil
    }

    // For unknown resources, try to discover them from the API server
    discoveryClient := clientset.Discovery()
    resources, err := discoveryClient.ServerPreferredResources()
    if err != nil {
        return schema.GroupVersionResource{}, fmt.Errorf("failed to get server resources: %v", err)
    }

    for _, list := range resources {
        for _, r := range list.APIResources {
            if r.Name == resourceType || r.SingularName == resourceType {
                gv, err := schema.ParseGroupVersion(list.GroupVersion)
                if err != nil {
                    continue
                }
                return schema.GroupVersionResource{
                    Group:    gv.Group,
                    Version:  gv.Version,
                    Resource: r.Name,
                }, nil
            }
        }
    }

    return schema.GroupVersionResource{}, fmt.Errorf("unknown resource type: %s", resourceType)
} 