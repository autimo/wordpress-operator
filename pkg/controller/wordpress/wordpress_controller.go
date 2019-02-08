package wordpress

import (
  "context"

  wordpressv1alpha1 "github.com/autimo/wordpress-operator/pkg/apis/wordpress/v1alpha1"

  appsv1 "k8s.io/api/apps/v1"
  corev1 "k8s.io/api/core/v1"
  "k8s.io/apimachinery/pkg/api/errors"
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  "k8s.io/apimachinery/pkg/runtime"
  "k8s.io/apimachinery/pkg/types"
  "k8s.io/apimachinery/pkg/util/intstr"
  "sigs.k8s.io/controller-runtime/pkg/client"
  "sigs.k8s.io/controller-runtime/pkg/controller"
  "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
  "sigs.k8s.io/controller-runtime/pkg/handler"
  "sigs.k8s.io/controller-runtime/pkg/manager"
  "sigs.k8s.io/controller-runtime/pkg/reconcile"
  logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
  "sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_wordpress")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Wordpress Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
  return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
  return &ReconcileWordpress{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
  // Create a new controller
  c, err := controller.New("wordpress-controller", mgr, controller.Options{Reconciler: r})
  if err != nil {
    return err
  }

  // Watch for changes to primary resource Wordpress
  err = c.Watch(&source.Kind{Type: &wordpressv1alpha1.Wordpress{}}, &handler.EnqueueRequestForObject{})
  if err != nil {
    return err
  }

  // TODO(user): Modify this to be the types you create that are owned by the primary resource
  // Watch for changes to secondary resource Pods and requeue the owner Wordpress
  err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
    IsController: true,
    OwnerType:    &wordpressv1alpha1.Wordpress{},
  })
  if err != nil {
    return err
  }

  return nil
}

var _ reconcile.Reconciler = &ReconcileWordpress{}

// ReconcileWordpress reconciles a Wordpress object
type ReconcileWordpress struct {
  // This client, initialized using mgr.Client() above, is a split client
  // that reads objects from the cache and writes to the apiserver
  client client.Client
  scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Wordpress object and makes changes based on the state read
// and what is in the Wordpress.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileWordpress) Reconcile(request reconcile.Request) (reconcile.Result, error) {
  reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
  reqLogger.Info("Reconciling Wordpress")

  // Fetch the Wordpress instance
  instance := &wordpressv1alpha1.Wordpress{}
  err := r.client.Get(context.TODO(), request.NamespacedName, instance)
  if err != nil {
    if errors.IsNotFound(err) {
      // Request object not found, could have been deleted after reconcile request.
      // Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
      // Return and don't requeue
      return reconcile.Result{}, nil
    }
    // Error reading the object - requeue the request.
    return reconcile.Result{}, err
  }

  // Define a new Service object
  service := newServiceForCR(instance)

  // Set Wordpress instance as the owner and controller
  if err = controllerutil.SetControllerReference(instance, service, r.scheme); err != nil {
    return reconcile.Result{}, err
  }

  if err = updateOrCreateService(r, instance, service); err != nil {
    return reconcile.Result{}, err
  }

  // Define a new DB Service object
  service = newDBServiceForCR(instance)

  // Set Wordpress instance as the owner and controller
  if err = controllerutil.SetControllerReference(instance, service, r.scheme); err != nil {
    return reconcile.Result{}, err
  }

  if err = updateOrCreateService(r, instance, service); err != nil {
    return reconcile.Result{}, err
  }

  // Define a new Pod object
  deploy := newDeployForCR(instance)

  // Set Wordpress instance as the owner and controller
  if err = controllerutil.SetControllerReference(instance, deploy, r.scheme); err != nil {
    return reconcile.Result{}, err
  }

  if err = updateOrCreateDeployment(r, instance, deploy); err != nil {
    return reconcile.Result{}, err
  }

  // Define a new DB Deployment object
  deploy = newDBDeployForCR(instance)

  // Set instance as the owner and controller
  if err = controllerutil.SetControllerReference(instance, deploy, r.scheme); err != nil {
    return reconcile.Result{}, err
  }

  if err = updateOrCreateDeployment(r, instance, deploy); err != nil {
    return reconcile.Result{}, err
  }
  return reconcile.Result{}, nil
}

func updateOrCreateService(r *ReconcileWordpress, instance *wordpressv1alpha1.Wordpress, service *corev1.Service) error {
  reqLogger := log.WithValues("Service.Namespace", service.Namespace, "Service.Name", service.Name)
  // Check if this Service already exists
  found := &corev1.Service{}
  err := r.client.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, found)
  if err != nil && errors.IsNotFound(err) {
    reqLogger.Info("Creating a new Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
    err = r.client.Create(context.TODO(), service)
    if err != nil {
      return err
    }
  } else if err != nil {
    return err
  } else {
    // Service already exists - update
    service.ObjectMeta.ResourceVersion = found.ObjectMeta.ResourceVersion
    service.Spec.ClusterIP = found.Spec.ClusterIP
    reqLogger.Info("Updating Service", "Service.Namespace", found.Namespace, "Service.Name", found.Name)
    err = r.client.Update(context.TODO(), service)
    if err != nil {
      return err
    }
  }
  return nil
}

func updateOrCreateDeployment(r *ReconcileWordpress, instance *wordpressv1alpha1.Wordpress, deploy *appsv1.Deployment) error {
  reqLogger := log.WithValues("Deployment.Namespace", deploy.Namespace, "Deployment.Name", deploy.Name)
  // Check if this Deployment already exists
  found := &appsv1.Deployment{}
  err := r.client.Get(context.TODO(), types.NamespacedName{Name: deploy.Name, Namespace: deploy.Namespace}, found)
  if err != nil && errors.IsNotFound(err) {
    reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", deploy.Namespace, "Deployment.Name", deploy.Name)
    err = r.client.Create(context.TODO(), deploy)
    if err != nil {
      return err
    }
  } else if err != nil {
    return err
  } else {
    // Deployment already exists - update
    deploy.ObjectMeta.ResourceVersion = found.ObjectMeta.ResourceVersion
    reqLogger.Info("Updating Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
    err = r.client.Update(context.TODO(), deploy)
    if err != nil {
      return err
    }
  }
  return nil
}

// newDeployForCR returns a wordpress service with the same name/namespace as the cr
func newServiceForCR(cr *wordpressv1alpha1.Wordpress) *corev1.Service {
  labels := map[string]string{
    "app": cr.Name,
  }

  port := cr.Spec.Port

  return &corev1.Service{
    ObjectMeta: metav1.ObjectMeta{
      Name:      cr.Name,
      Namespace: cr.Namespace,
      Labels:    labels,
    },
    Spec: corev1.ServiceSpec{
      Ports: []corev1.ServicePort{
        {Name: "http", Port: port, Protocol: "TCP", TargetPort: intstr.FromInt(80)},
      },
      Selector:        labels,
      SessionAffinity: corev1.ServiceAffinityNone,
      Type:            corev1.ServiceTypeLoadBalancer,
    },
  }
}

// newDeployForCR returns a wordpress deployment with the same name/namespace as the cr
func newDeployForCR(cr *wordpressv1alpha1.Wordpress) *appsv1.Deployment {
  labels := map[string]string{
    "app": cr.Name,
  }

  version := cr.Spec.WordpressVersion
  replicas := cr.Spec.Size

  return &appsv1.Deployment{
    TypeMeta: metav1.TypeMeta{
      APIVersion: "apps/v1",
      Kind:       "Deployment",
    },
    ObjectMeta: metav1.ObjectMeta{
      Name:      cr.Name,
      Namespace: cr.Namespace,
    },
    Spec: appsv1.DeploymentSpec{
      Replicas: &replicas,
      Selector: &metav1.LabelSelector{
        MatchLabels: labels,
      },
      Template: corev1.PodTemplateSpec{
        ObjectMeta: metav1.ObjectMeta{
          Labels: labels,
        },
        Spec: corev1.PodSpec{
          Containers: []corev1.Container{
            {
              Name:  "wordpress",
              Image: "wordpress:" + version,
              Ports: []corev1.ContainerPort{{
                ContainerPort: 80,
                Name:          "wordpress-http",
              }},
              Env: []corev1.EnvVar{
                {
                  Name:  "WORDPRESS_DB_HOST",
                  Value: cr.Name + "-db",
                },
                {
                  Name:  "WORDPRESS_DB_PASSWORD",
                  Value: cr.Spec.DBPassword,
                },
                {
                  Name:  "WORDPRESS_DB_USER",
                  Value: cr.Spec.DBUsername,
                },
                {
                  Name:  "WORDPRESS_DB_NAME",
                  Value: cr.Spec.DBName,
                },
              },
            },
          },
        },
      },
    },
  }
}

// newDeployForCR returns a wordpress service with the same name/namespace as the cr
func newDBServiceForCR(cr *wordpressv1alpha1.Wordpress) *corev1.Service {
  labels := map[string]string{
    "app": cr.Name + "-db",
  }

  return &corev1.Service{
    ObjectMeta: metav1.ObjectMeta{
      Name:      cr.Name + "-db",
      Namespace: cr.Namespace,
      Labels:    labels,
    },
    Spec: corev1.ServiceSpec{
      Ports: []corev1.ServicePort{
        {Name: "http", Port: 3306, Protocol: "TCP", TargetPort: intstr.FromInt(3306)},
      },
      Selector:        labels,
      SessionAffinity: corev1.ServiceAffinityNone,
      Type:            corev1.ServiceTypeClusterIP,
    },
  }
}

// newDBDeployForCR returns a mysql deployment with the same name/namespace as the cr
func newDBDeployForCR(cr *wordpressv1alpha1.Wordpress) *appsv1.Deployment {
  labels := map[string]string{
    "app": cr.Name + "-db",
  }

  version := cr.Spec.MysqlVersion
  var replicas int32
  replicas = 1

  return &appsv1.Deployment{
    TypeMeta: metav1.TypeMeta{
      APIVersion: "apps/v1",
      Kind:       "Deployment",
    },
    ObjectMeta: metav1.ObjectMeta{
      Name:      cr.Name + "-db",
      Namespace: cr.Namespace,
    },
    Spec: appsv1.DeploymentSpec{
      Replicas: &replicas,
      Selector: &metav1.LabelSelector{
        MatchLabels: labels,
      },
      Template: corev1.PodTemplateSpec{
        ObjectMeta: metav1.ObjectMeta{
          Labels: labels,
        },
        Spec: corev1.PodSpec{
          Containers: []corev1.Container{
            {
              Name:  "wordpress-db",
              Image: "mysql:" + version,
              Ports: []corev1.ContainerPort{{
                ContainerPort: 3306,
                Name:          "mysql",
              }},
              Env: []corev1.EnvVar{
                {
                  Name:  "MYSQL_ROOT_PASSWORD",
                  Value: cr.Spec.DBPassword,
                },
                {
                  Name:  "MYSQL_PASSWORD",
                  Value: cr.Spec.DBPassword,
                },
                {
                  Name:  "MYSQL_USER",
                  Value: cr.Spec.DBUsername,
                },
                {
                  Name:  "MYSQL_DATABASE",
                  Value: cr.Spec.DBName,
                },
              },
            },
          },
        },
      },
    },
  }
}
