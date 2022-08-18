/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"math/rand"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	testingv1 "chaos.io/testing/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// PodChaosMonkeyReconciler reconciles a PodChaosMonkey object
type PodChaosMonkeyReconciler struct {
	kubeclient    kubernetes.Interface
	eventRecorder record.EventRecorder
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=testing.chaos.io,resources=podchaosmonkeys,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=testing.chaos.io,resources=podchaosmonkeys/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=testing.chaos.io,resources=podchaosmonkeys/finalizers,verbs=update

// Get List of PodChaosMonkey Objects
func (r *PodChaosMonkeyReconciler) GetPodChaosMonkeyList(ctx context.Context, name, namespace, kind, apiVersion string) *testingv1.PodChaosMonkeyList {
	chaosList := &testingv1.PodChaosMonkeyList{}

	lo := (&client.ListOptions{}).ApplyOptions([]client.ListOption{
		client.MatchingFields{"spec.targetRef.kind": kind},
		client.MatchingFields{"spec.targetRef.name": name},
		client.MatchingFields{"spec.targetRef.apiVersion": apiVersion},
		client.InNamespace(namespace),
	})
	err := r.Client.List(context.TODO(), chaosList, lo)
	if err != nil {
		lo = (&client.ListOptions{}).ApplyOptions([]client.ListOption{
			client.MatchingFields{"spec.targetRef.kind": kind},
			client.MatchingFields{"spec.targetRef.name": name},
			client.InNamespace(namespace),
		})
		err := r.Client.List(context.TODO(), chaosList, lo)
		if err != nil {
			lo = (&client.ListOptions{}).ApplyOptions([]client.ListOption{
				client.MatchingFields{"spec.targetRef.kind": kind},
				client.MatchingFields{"spec.targetRef.apiVersion": apiVersion},
				client.InNamespace(namespace),
			})
			err := r.Client.List(context.TODO(), chaosList, lo)
			if err != nil {
				return nil
			}
		}
	}
	return chaosList
}

// Get List of ReplicaSet Objects
func (r *PodChaosMonkeyReconciler) GetListReplicaSet(ctx context.Context, name, namespace, kind, apiVersion string) []appsv1.ReplicaSet {
	returnList := []appsv1.ReplicaSet{}
	repList := &appsv1.ReplicaSetList{}
	lo := (&client.ListOptions{}).ApplyOptions([]client.ListOption{
		client.InNamespace(namespace),
	})
	err := r.Client.List(context.TODO(), repList, lo)
	if err != nil {
		return returnList
	}
	for _, rep := range repList.Items {
		if rep.ObjectMeta.OwnerReferences[0].Name == name && rep.ObjectMeta.OwnerReferences[0].Kind == kind && rep.ObjectMeta.OwnerReferences[0].APIVersion == apiVersion &&
			rep.Status.ReadyReplicas > 0 {
			returnList = append(returnList, rep)
		}
	}
	return returnList
}

// Get List of Pod Objects
func (r *PodChaosMonkeyReconciler) GetListPodsRunning(ctx context.Context, name, namespace, kind, apiVersion string) []corev1.Pod {
	returnList := []corev1.Pod{}
	podList := &corev1.PodList{}
	lo := (&client.ListOptions{}).ApplyOptions([]client.ListOption{
		client.InNamespace(namespace),
	})
	err := r.Client.List(context.TODO(), podList, lo)
	if err != nil {
		return returnList
	}
	for _, pod := range podList.Items {
		if pod.ObjectMeta.OwnerReferences[0].Kind == "ReplicaSet" {
			repList := r.GetListReplicaSet(ctx, name, namespace, kind, apiVersion)
			for _, rep := range repList {
				if rep.GetName() == pod.ObjectMeta.OwnerReferences[0].Name &&
					rep.Kind == pod.ObjectMeta.OwnerReferences[0].Kind &&
					rep.APIVersion == pod.ObjectMeta.OwnerReferences[0].APIVersion &&
					pod.Status.Phase == "Running" {
					returnList = append(returnList, pod)
				}
			}
		} else {
			if pod.ObjectMeta.OwnerReferences[0].Name == name && pod.ObjectMeta.OwnerReferences[0].Kind == kind &&
				pod.ObjectMeta.OwnerReferences[0].APIVersion == apiVersion && pod.Status.Phase == "Running" {
				returnList = append(returnList, pod)
			}
		}
	}
	return returnList
}

// Reconcile Object
// Find and process a PodChaosMonkey Object, that matchs kind of object
func (r *PodChaosMonkeyReconciler) ReconcileObject(ctx context.Context, req ctrl.Request, name, namespace, kind, apiVersion string) (ctrl.Result, error) {
	// log := log.FromContext(ctx)

	chaosList := r.GetPodChaosMonkeyList(ctx, name, namespace, kind, apiVersion)
	if chaosList != nil {
		for _, chaos := range chaosList.Items {
			ret, err := r.ReconcileChaos(ctx, req, chaos, name, namespace, kind, apiVersion)
			if err != nil {
				return ret, err
			}
		}
	}

	return ctrl.Result{}, nil
}

// Reconcile Object
// Find and process a PodChaosMonkey Object, that matchs kind of object
func (r *PodChaosMonkeyReconciler) ReconcileChaos(ctx context.Context, req ctrl.Request, chaos testingv1.PodChaosMonkey, name, namespace, kind, apiVersion string) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	podList := r.GetListPodsRunning(ctx, name, namespace, kind, apiVersion)
	if int32(len(podList)) >= *chaos.Spec.Conditions.MinPods {
		newList := []corev1.Pod{}
		currTime := time.Now()
		for _, pod := range podList {
			duration := currTime.Sub(pod.ObjectMeta.CreationTimestamp.Time)
			if int32(duration.Minutes()) >= *chaos.Spec.Conditions.MinRunning {
				newList = append(newList, pod)
			}
		}
		if len(newList) > 0 {
			random := rand.Intn(len(newList))
			podToEvict := newList[random]
			log.Info("Reconciling PodChaosMonkey", "Process", req.Name, "Going to Evict Pod", podToEvict.ObjectMeta.Name)

			eviction := &policyv1.Eviction{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: podToEvict.Namespace,
					Name:      podToEvict.Name,
				},
			}
			err := r.kubeclient.CoreV1().Pods(podToEvict.Namespace).Evict(context.TODO(), eviction)
			if err != nil {
				log.Error(err, "Failed to Evict Pod", "Pod", podToEvict.ObjectMeta.Name)
				return ctrl.Result{}, err
			}
			r.eventRecorder.Event(&podToEvict, apiv1.EventTypeNormal, "EvictedByChaosMonkey",
				"Pod was evicted by Chaos Monkey.")
		}
	}
	return ctrl.Result{}, nil
}

// Reconcile DaemonSet
// Find and process a PodChaosMonkey Object, that matchs kind of object
func (r *PodChaosMonkeyReconciler) ReconcileDaemonSet(ctx context.Context, req ctrl.Request, d *appsv1.DaemonSet) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	log.Info("Reconciling PodChaosMonkey", "Process DaemonSet", req.Name)
	return r.ReconcileObject(ctx, req, d.ObjectMeta.Name, d.ObjectMeta.Namespace, d.Kind, d.APIVersion)
}

// Reconcile StatefulSet
// Find and process a PodChaosMonkey Object, that matchs kind of object
func (r *PodChaosMonkeyReconciler) ReconcileStatefulSet(ctx context.Context, req ctrl.Request, st *appsv1.StatefulSet) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Reconciling PodChaosMonkey", "Process StatefulSet", req.Name)
	return r.ReconcileObject(ctx, req, st.ObjectMeta.Name, st.ObjectMeta.Namespace, st.Kind, st.APIVersion)
}

// Reconcile Deployment
// Find and process a PodChaosMonkey Object, that matchs kind of object
func (r *PodChaosMonkeyReconciler) ReconcileDeployment(ctx context.Context, req ctrl.Request, dep *appsv1.Deployment) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Reconciling PodChaosMonkey", "Process Deployment", req.Name)
	repList := r.GetListReplicaSet(ctx, dep.GetName(), dep.GetNamespace(), dep.Kind, dep.APIVersion)
	for _, rep := range repList {
		ret, err := r.ReconcileObject(ctx, req, rep.ObjectMeta.Name, rep.ObjectMeta.Namespace, rep.Kind, rep.APIVersion)
		if err != nil {
			return ret, err
		}
	}
	return ctrl.Result{}, nil
}

// Reconcile PodChaosMonkey
// Find and process a PodChaosMonkey Object, that matchs kind of object
func (r *PodChaosMonkeyReconciler) ReconcilePodChaosMonkey(ctx context.Context, req ctrl.Request, chaos testingv1.PodChaosMonkey) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Reconciling PodChaosMonkey", "Process PodChaosMonkey", req.Name)

	objList := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"kind":       chaos.Spec.TargetRef.Kind,
			"apiVersion": chaos.Spec.TargetRef.APIVersion,
		},
	}
	lo := (&client.ListOptions{}).ApplyOptions([]client.ListOption{
		//client.MatchingFields{"metadata.name": vpaName},
		client.InNamespace(chaos.ObjectMeta.Namespace),
	})
	err := r.Client.List(context.TODO(), objList, lo)
	if err != nil {
		log.Error(err, "Failed to list objects", "Kind", chaos.Spec.TargetRef.Kind, "Namespace", chaos.ObjectMeta.Namespace)
		return ctrl.Result{}, err
	}
	for _, obj := range objList.Items {
		if chaos.Spec.TargetRef.Name == "" || chaos.Spec.TargetRef.Name == obj.GetName() {
			ret, err := r.ReconcileChaos(ctx, req, chaos, obj.GetName(), obj.GetNamespace(), obj.GetKind(), obj.GetAPIVersion())
			if err != nil {
				return ret, err
			}
		}
	}

	return ctrl.Result{}, nil
}

// Reconcile main function is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *PodChaosMonkeyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the AutomaticVPA from the cache
	chaosMonkey := &testingv1.PodChaosMonkey{}
	err := r.Client.Get(ctx, req.NamespacedName, chaosMonkey)
	if errors.IsNotFound(err) || err != nil {
		obj := &appsv1.Deployment{}
		err = r.Client.Get(ctx, req.NamespacedName, obj)
		if errors.IsNotFound(err) && err != nil {
			obj := &appsv1.DaemonSet{}
			err = r.Client.Get(ctx, req.NamespacedName, obj)
			if errors.IsNotFound(err) && err != nil {
				obj := &appsv1.StatefulSet{}
				err = r.Client.Get(ctx, req.NamespacedName, obj)
				if errors.IsNotFound(err) && err != nil {
					log.Info("Reconciling PodChaosMonkey", "Error", req.Name)
					return ctrl.Result{}, err
				} else {
					// Reconcile StatefulSet
					return r.ReconcileStatefulSet(ctx, req, obj)
				}
			} else {
				// Process DaemonSet
				return r.ReconcileDaemonSet(ctx, req, obj)
			}
		} else {
			// Process Deployment
			return r.ReconcileDeployment(ctx, req, obj)
		}
	}
	log.Info("Reconciling PodChaosMonkey", "Process PodChaosMonkey", req.Name)
	return r.ReconcilePodChaosMonkey(ctx, req, *chaosMonkey)
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodChaosMonkeyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	//TODO set kubeconfig
	return ctrl.NewControllerManagedBy(mgr).
		For(&testingv1.PodChaosMonkey{}).
		Watches(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForObject{}).
		Watches(&source.Kind{Type: &appsv1.DaemonSet{}}, &handler.EnqueueRequestForObject{}).
		Watches(&source.Kind{Type: &appsv1.StatefulSet{}}, &handler.EnqueueRequestForObject{}).
		Complete(r)
}
