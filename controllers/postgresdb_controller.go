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
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	dbv1 "github.com/heinrichgrt/pg-operator/api/v1"
)

// PostgresDBReconciler reconciles a PostgresDB object
type PostgresDBReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}
type postgresParent struct {
	name     string
	password string
	username string
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PostgresDB object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func connectToDB(key string) (*pgx.Conn, error) {
	//todo make this more robust
	//tood pgCredentials-map does not work. Replace with own code...
	//DATABASE_URL := "postgres://" + pgCredentials[key].Username + ":" + pgCredentials[key].Password + "@" + pgCredentials[key].ServerURL + "/postgres"
	DATABASE_URL := "postgres://" + pgCredentials.Username + ":" + pgCredentials.Password + "@" + pgCredentials.ServerURL + "/postgres"
	conn, err := pgx.Connect(context.Background(), DATABASE_URL)
	fmt.Printf("trying to connect %v\n", DATABASE_URL)
	if err == nil {
		//defer conn.Close(context.Background())
		fmt.Printf("connected to  %v\n", DATABASE_URL)
	}
	return conn, err
}
func doesDBExist(conn *pgx.Conn, dbname string) (bool, error) {
	// todo that looks ugly...
	STATEMENT := "SELECT 1 FROM pg_database WHERE datname='" + dbname + "';"
	rows, err := conn.Query(context.
		Background(), STATEMENT)
	if err != nil {
		return false, err
	}
	count := 0
	for rows.Next() {
		count++
	}
	if count == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

func createRole(conn *pgx.Conn, roleName string) (bool, error) {

}

//func createDBandRole (conn *pgx.Conn, dbname string, role string) (bool, error){
//
//	return true, nil
//}
func doesDBExistAndIsOwnerRight(h *pgx.Conn, dbname string, owner string) (bool, error) {
	//todo complete
	STATEMENT := "SELECT d.datname as \"Name\",pg_catalog.pg_get_userbyid(d.datdba) as \"Owner\" FROM pg_catalog.pg_database d WHERE d.datname = " + dbname
	fmt.Printf("%v", STATEMENT)
	return true, nil
}

//+kubebuilder:rbac:groups=db.grotjohann.com,resources=postgresdbs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=db.grotjohann.com,resources=postgresdbs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=db.grotjohann.com,resources=postgresdbs/finalizers,verbs=update
func (r *PostgresDBReconciler) ConnectToDB(ctx context.Context, req ctrl.Request) (res *postgresParent, error) {
	res = postgresParent{}
	postgresdb := &dbv1.PostgresDB{}
	err := r.Get(ctx, req.NamespacedName, postgresdb)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Postgres resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Postgres")
		//return ctrl.Result{}, err
	}
	return res, err
}

func (r *PostgresDBReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("DB: Start reconcile DB Controller")
	postgresdb := &dbv1.PostgresDB{}
	err := r.Get(ctx, req.NamespacedName, postgresdb)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Postgres resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Postgres")
		return ctrl.Result{}, err
	}
	// fetch parent:
	postgres := &dbv1.Postgres{}
	namespaceName := types.NamespacedName{
		Namespace: "",
		Name:      postgresdb.Spec.ParentDB,
	}
	err = r.Get(ctx, namespaceName, postgres)
	// fetch the secret:
	secret := &corev1.Secret{}
	namespaceName = types.NamespacedName{
		Namespace: "pg-operator-system",
		Name:      postgres.Spec.DbSecret,
	}
	r.Client.Get(ctx, namespaceName, secret)
	// make use of it:
	pgcredentials(string(secret.Data["secret.txt"]), postgres)

	log.Info("DB: we are in businiss")
	// need to fetch the parent Spec
	dbhande, err := connectToDB(postgresdb.Spec.ParentDB)
	if err != nil {
		log.Info("cannot connect to db")
		return ctrl.Result{}, nil
	}
	defer dbhande.Close(context.Background())
	log.Info("DB: connected to db")

	res, err := doesDBExist(dbhande, postgresdb.Name)
	if res != true {
		log.Info("DB: " + postgresdb.Name + " does not exist! ")

	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PostgresDBReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dbv1.PostgresDB{}).
		Complete(r)
}
