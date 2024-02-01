package controller

import (
	"context"
	"fmt"

	immudbiov1 "github.com/MathieuCesbron/immudb-operator/api/v1"
	"github.com/MathieuCesbron/immudb-operator/internal/controller/common"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/types"
)

func (r *ImmudbReconciler) CreateDatabase(ctx context.Context, immudb *immudbiov1.Immudb) error {
	sts := &appsv1.StatefulSet{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: immudb.Namespace,
		Name:      immudb.Name,
	}, sts)

	// create if statefulset does not exist
	if err != nil && k8serrors.IsNotFound(err) {
		sts := r.GetStatefulset(immudb)
		r.Log.Info("creating statefulset")
		err := r.Create(ctx, sts)
		if err != nil {
			return fmt.Errorf("error creating statefulset: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("error getting statefulset: %w", err)
	} else {
		// update if statefulset config is wrong
		if *sts.Spec.Replicas != *immudb.Spec.Replicas {
			sts := r.GetStatefulset(immudb)
			r.Log.Info("updating statefulset replicas field")
			err = r.Update(ctx, sts)
			if err != nil {
				return fmt.Errorf("error updating statefulset replicas field: %w", err)
			}
		}
	}

	return nil
}

func (r *ImmudbReconciler) GetStatefulset(immudb *immudbiov1.Immudb) *appsv1.StatefulSet {
	ls := common.GetLabels(immudb.Name)
	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:            immudb.Name,
			Namespace:       immudb.Namespace,
			OwnerReferences: common.GetOwnerReferences(immudb),
			Labels:          common.GetLabels(immudb.Name),
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: immudb.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: "codenotary/immudb:1.2.2",
							Name:  "immudb",
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 8080,
								},
								{
									Name:          "grpc",
									ContainerPort: 3322,
								},
								{
									Name:          "metrics",
									ContainerPort: 9497,
								},
							},
						},
					},
				},
			},
		},
	}
}
