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
	"sync"
	"time"

	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	testingv1 "chaos.io/testing/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
)

// PodChaosMonkeyReconciler reconciles a PodChaosMonkey object
type PodChaosMonkeyReconciler struct {
	record.EventRecorder
	KubeClient kubernetes.Interface
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
	err := r.Client.List(ctx, chaosList, lo)
	if err != nil {
		lo = (&client.ListOptions{}).ApplyOptions([]client.ListOption{
			client.MatchingFields{"spec.targetRef.kind": kind},
			client.MatchingFields{"spec.targetRef.name": name},
			client.InNamespace(namespace),
		})
		err := r.Client.List(ctx, chaosList, lo)
		if err != nil {
			lo = (&client.ListOptions{}).ApplyOptions([]client.ListOption{
				client.MatchingFields{"spec.targetRef.kind": kind},
				client.MatchingFields{"spec.targetRef.apiVersion": apiVersion},
				client.InNamespace(namespace),
			})
			err := r.Client.List(ctx, chaosList, lo)
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
	err := r.Client.List(ctx, repList, lo)
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
	err := r.Client.List(ctx, podList, lo)
	if err != nil {
		return returnList
	}

	for _, pod := range podList.Items {
		if pod.ObjectMeta.OwnerReferences != nil {
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
	}
	return returnList
}

// Reconcile Object
// Find and process a PodChaosMonkey Object, that matchs kind of object
func (r *PodChaosMonkeyReconciler) ReconcileObject(ctx context.Context, name, namespace, kind, apiVersion string) (ctrl.Result, error) {
	// log := log.FromContext(ctx)

	chaosList := r.GetPodChaosMonkeyList(ctx, name, namespace, kind, apiVersion)
	if chaosList != nil {
		for _, chaos := range chaosList.Items {
			ret, err := r.ReconcileChaos(ctx, chaos, name, namespace, kind, apiVersion)
			if err != nil {
				return ret, err
			}
		}
	}

	return ctrl.Result{}, nil
}

// Reconcile Object
// Find and process a PodChaosMonkey Object, that matchs kind of object
func (r *PodChaosMonkeyReconciler) ReconcileChaos(ctx context.Context, chaos testingv1.PodChaosMonkey, name, namespace, kind, apiVersion string) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	podList := r.GetListPodsRunning(ctx, name, namespace, kind, apiVersion)
	if int32(len(podList)) >= *chaos.Spec.Conditions.MinPods {
		newList := []corev1.Pod{}
		for _, pod := range podList {
			if pod.Status.ContainerStatuses[0].State.Running != nil {
				currTime := time.Now()
				duration := currTime.Sub(pod.Status.ContainerStatuses[0].State.Running.StartedAt.Time)
				x := duration.Minutes()
				var y float64 = 0.0
				if chaos.Status.LastEvictionAt != nil {
					currTime = time.Now()
					y = currTime.Sub(chaos.Status.LastEvictionAt.Time).Minutes()
				}
				durationBetweenEvictions := float64(rand.Intn(int(*chaos.Spec.Conditions.MaxTimeRandom-*chaos.Spec.Conditions.MinTimeRandom)) + int(*chaos.Spec.Conditions.MinTimeRandom))
				if int32(x) >= *chaos.Spec.Conditions.MinTimeRunning && (chaos.Status.LastEvictionAt == nil || y >= durationBetweenEvictions) {
					newList = append(newList, pod)
				}
			}
		}
		if len(newList) > 0 {
			random := rand.Intn(len(newList))
			podToEvict := newList[random]
			do := &client.DeleteOptions{}
			client.GracePeriodSeconds(1).ApplyToDelete(do)
			log.Info("Reconciling PodChaosMonkey", "Process", name, "Going to Evict Pod", podToEvict.ObjectMeta.Name)
			var gracePeriodSeconds int64 = 1
			if chaos.Spec.Conditions.GracePeriodSeconds != nil {
				gracePeriodSeconds = *chaos.Spec.Conditions.GracePeriodSeconds
			}
			eviction := &policyv1beta1.Eviction{
				TypeMeta: metav1.TypeMeta{
					APIVersion: podToEvict.APIVersion,
					Kind:       podToEvict.Kind,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      podToEvict.ObjectMeta.Name,
					Namespace: podToEvict.ObjectMeta.Namespace,
				},
				DeleteOptions: &metav1.DeleteOptions{
					GracePeriodSeconds: &gracePeriodSeconds,
				},
			}

			err := r.KubeClient.PolicyV1beta1().Evictions(eviction.Namespace).Evict(ctx, eviction)
			//err := r.Client.Delete(ctx, &podToEvict, do)
			if err != nil {
				log.Error(err, "Failed to Evict Pod", "Pod", podToEvict.ObjectMeta.Name)
				return ctrl.Result{}, err
			}
			chaos.Status.LastEvictionAt = &metav1.Time{Time: time.Now()}
			var num int32 = 1
			if chaos.Status.NumberOfEvictions != nil {
				num = *chaos.Status.NumberOfEvictions + 1
			}
			chaos.Status.NumberOfEvictions = &num
			r.Client.Status().Update(ctx, &chaos)
			log.Info("Reconciling PodChaosMonkey", "Going to Record Event", podToEvict.ObjectMeta.Name)
			r.EventRecorder.Event(&podToEvict, apiv1.EventTypeNormal, "EvictedByChaosMonkey",
				"Pod was evicted by Chaos Monkey.")
		}
	}
	return ctrl.Result{}, nil
}

// Reconcile PodChaosMonkey
// Find and process a PodChaosMonkey Object, that matchs kind of object
func (r *PodChaosMonkeyReconciler) ReconcilePodChaosMonkey(ctx context.Context, chaos testingv1.PodChaosMonkey) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Reconciling PodChaosMonkey", "Process PodChaosMonkey", chaos.ObjectMeta.Name)

	objList := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"kind":       chaos.Spec.TargetRef.Kind,
			"apiVersion": chaos.Spec.TargetRef.APIVersion,
		},
	}
	lo := (&client.ListOptions{}).ApplyOptions([]client.ListOption{
		//client.MatchingFields{"metadata.name": chaos.ObjectMeta.Name}
		client.InNamespace(chaos.ObjectMeta.Namespace),
	})
	err := r.Client.List(ctx, objList, lo)
	if err != nil {
		log.Error(err, "Failed to list objects", "Kind", chaos.Spec.TargetRef.Kind, "Namespace", chaos.ObjectMeta.Namespace)
		return ctrl.Result{}, err
	}
	for _, obj := range objList.Items {
		if chaos.Spec.TargetRef.Name == "" || chaos.Spec.TargetRef.Name == obj.GetName() {
			ret, err := r.ReconcileChaos(ctx, chaos, obj.GetName(), obj.GetNamespace(), obj.GetKind(), obj.GetAPIVersion())
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

	log.Info("Reconciling PodChaosMonkey", "Process PodChaosMonkey", req.Name)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodChaosMonkeyReconciler) Runnable(mgr ctrl.Manager, wg *sync.WaitGroup) {
	defer wg.Done()
	ctx := context.TODO()
	log := log.FromContext(ctx)

	// while true loop
	for {
		log.Info("Reconciling PodChaosMonkey", "Runnable PodChaosMonkey", "Get PodChaosMonkey List")
		chaosList := &testingv1.PodChaosMonkeyList{}
		lo := (&client.ListOptions{}).ApplyOptions([]client.ListOption{})
		err := r.Client.List(ctx, chaosList, lo)
		if err != nil {
			log.Error(err, "Failed to list objects", "Kind", "ChaosList", "Namespace", "Error")
			continue
		}
		for _, chaos := range chaosList.Items {
			_, err := r.ReconcilePodChaosMonkey(ctx, chaos)
			if err != nil {
				log.Error(err, "Failed to ReconcilePodChaosMonkey", "Kind", "ChaosList", "Namespace", chaos.ObjectMeta.Name)
				continue
			}
		}
		time.Sleep(10000 * time.Millisecond)
	}

}

// SetupWithManager sets up the controller with the Manager.
func (r *PodChaosMonkeyReconciler) SetupWithManager(mgr ctrl.Manager) error {

	var wg sync.WaitGroup
	wg.Add(1)
	go r.Runnable(mgr, &wg)

	return ctrl.NewControllerManagedBy(mgr).
		For(&testingv1.PodChaosMonkey{}).
		Complete(r)

}
