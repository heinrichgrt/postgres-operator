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
	dbv1 "github.com/heinrichgrt/pg-operator/api/v1"
	"github.com/jackc/pgx/v4"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"
	//"sigs.k8s.io/controller-runtime/pkg/log"
	"fmt"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// PostgresReconciler reconciles a Postgres object
type PostgresReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}
type PostgresCredential struct {
	Username  string
	Password  string
	ServerURL string
	Created   time.Time
}

var (
	pgCredentials = map[string]PostgresCredential{}
)

func pgcredentials(c string, p *dbv1.Postgres) {
	// todo: make this nice please...
	res := make(map[string]string)
	b := strings.Split(c, "\n")
	for s := range b {
		if len(b[s]) < 1 {
			break
		}
		d := strings.Split(b[s], ":")
		res[strings.TrimSpace(d[0])] = strings.TrimSpace(d[1])
	}
	pg := new(PostgresCredential)
	pg.Username = res["user"]
	pg.Password = res["pass"]
	pg.Created = time.Now()
	pg.ServerURL = p.Spec.ConnectionUrl
	if _, ok := pgCredentials[p.Name]; !ok {
		pgCredentials[p.Name] = *pg
	}
}

func checkConnection(s *dbv1.Postgres) error {
	DATABASE_URL := "postgres://" + pgCredentials[s.Name].Username + ":" + pgCredentials[s.Name].Password + "@" + s.Spec.ConnectionUrl + "/postgres"
	conn, err := pgx.Connect(context.Background(), DATABASE_URL)
	fmt.Printf("trying to connect %v\n", DATABASE_URL)
	if err == nil {
		defer conn.Close(context.Background())
		fmt.Printf("connected to  %v\n", DATABASE_URL)
	}

	return err
}

//+kubebuilder:rbac:groups=db.grotjohann.com,resources=postgres,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=db.grotjohann.com,resources=postgres/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=db.grotjohann.com,resources=postgres/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Postgres object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *PostgresReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Starting reconcile")
	postgres := &dbv1.Postgres{}
	err := r.Get(ctx, req.NamespacedName, postgres)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Postgres resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Postgres")
		return ctrl.Result{}, err
	}
	// we need the secret for this user
	//todo this will not fail, if the secret does not exist, why?
	secret := &corev1.Secret{}
	namespaceName := types.NamespacedName{
		Namespace: "pg-operator-system",
		Name:      postgres.Spec.DbSecret,
	}
	r.Client.Get(ctx, namespaceName, secret)

	if err != nil {
		log.Error(err, "could not fetch secret")
		//todo add some status for missing secret
		// if the secret is not there it should throw an error, but it does not?
		return ctrl.Result{}, err
	}
	ConnectionIsPossible := true
	pgcredentials(string(secret.Data["secret.txt"]), postgres)
	log.Info("still working?")
	log.Info("will check " + postgres.Spec.ConnectionUrl)
	err = checkConnection(postgres)
	if err != nil {
		ConnectionIsPossible = false
		log.Error(err, "can not connect to db")
		if !reflect.DeepEqual(ConnectionIsPossible, postgres.Status.Reachable) {
			postgres.Status.Reachable = ConnectionIsPossible
			postgres.Status.LastTry = &metav1.Time{Time: time.Now()}
			err := r.Status().Update(ctx, postgres)
			if err != nil {
				log.Error(err, "Failed to update pgdatabase status")
				return ctrl.Result{}, err
			}
		}
		//todo check the return code!
		return ctrl.Result{}, err
	}
	// update the status
	postgres.Status.Reachable = ConnectionIsPossible
	postgres.Status.LastTry = &metav1.Time{Time: time.Now()}
	postgres.Status.LastConnect = postgres.Status.LastTry
	err = r.Status().Update(ctx, postgres)
	if err != nil {
		log.Error(err, "Failed to update pgdatabase status")
		return ctrl.Result{}, err
	}

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PostgresReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dbv1.Postgres{}).
		Complete(r)
}
