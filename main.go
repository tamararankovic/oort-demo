package main

import (
	"context"
	"log"
	"time"

	oort "github.com/c12s/oort/pkg/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	dial, err := grpc.Dial("localhost:8000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalln(err)
	}
	administrator := oort.NewOortAdministratorClient(dial)
	evaluator := oort.NewOortEvaluatorClient(dial)

	administratorAsync, err := oort.NewAdministrationAsyncClient("localhost:4222")
	if err != nil {
		log.Fatalln(err)
	}

	basicRPCs(administrator, administratorAsync, evaluator)
	grantedPermissions(evaluator)
}

func basicRPCs(administrator oort.OortAdministratorClient, administratorAsync *oort.AdministrationAsyncClient, evaluator oort.OortEvaluatorClient) {
	// svi unutar grupe mogu da citaju konfiguracije unutar roditeljskog ns-a
	// zakomentarisani deo salje preko grpc klijenta, a ispod je asinhrono slanje zahteva
	//_, err = administrator.CreatePolicy(context.TODO(), &oort.CreatePolicyReq{
	//	SubjectScope: group,
	//	ObjectScope:  parentNamespace,
	//	Permission:   getConfigPerm,
	//})
	err := administratorAsync.SendRequest(&oort.CreatePolicyReq{
		SubjectScope: group,
		ObjectScope:  parentNamespace,
		Permission:   getConfigPerm,
	}, func(resp *oort.AdministrationAsyncResp) {
		if len(resp.Error) > 0 {
			log.Println(resp.Error)
		}
	})
	if err != nil {
		log.Fatalln(err)
	}
	// sve unutar ns-a (sve aplikacije) mogu da citaju konfiguraciju unutar tog ns-a
	_, err = administrator.CreatePolicy(context.TODO(), &oort.CreatePolicyReq{
		SubjectScope: parentNamespace,
		ObjectScope:  parentNamespace,
		Permission:   getConfigPerm,
	})
	if err != nil {
		log.Fatalln(err)
	}
	_, err = administrator.CreatePolicy(context.TODO(), &oort.CreatePolicyReq{
		SubjectScope: childNamespace,
		ObjectScope:  childNamespace,
		Permission:   getConfigPerm,
	})
	if err != nil {
		log.Fatalln(err)
	}
	// aplikacije iz ns-a ne mogu da citaju konfiguracije roditeljskih ns-ova
	_, err = administrator.CreatePolicy(context.TODO(), &oort.CreatePolicyReq{
		SubjectScope: childNamespace,
		ObjectScope:  parentNamespace,
		Permission:   denyGetConfigPerm,
	})
	if err != nil {
		log.Fatalln(err)
	}
	// korisnici nasledjuje dozvole iz grupe
	// _, err = administrator.CreateInheritanceRel(context.TODO(), &oort.CreateInheritanceRelReq{
	// 	From: group,
	// 	To:   user1,
	// })
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	err = administratorAsync.SendRequest(&oort.CreateInheritanceRelReq{
		From: group,
		To:   user1,
	}, func(resp *oort.AdministrationAsyncResp) {
		if len(resp.Error) > 0 {
			log.Println(resp.Error)
		}
	})
	if err != nil {
		log.Fatalln(err)
	}
	_, err = administrator.CreateInheritanceRel(context.TODO(), &oort.CreateInheritanceRelReq{
		From: group,
		To:   user2,
	})
	if err != nil {
		log.Fatalln(err)
	}
	// child ns nasledjuje dozvole od roditeljskog ns-a
	_, err = administrator.CreateInheritanceRel(context.TODO(), &oort.CreateInheritanceRelReq{
		From: parentNamespace,
		To:   childNamespace,
	})
	if err != nil {
		log.Fatalln(err)
	}
	// svaka konfiguracija i aplikacija pripada ns-u
	_, err = administrator.CreateInheritanceRel(context.TODO(), &oort.CreateInheritanceRelReq{
		From: parentNamespace,
		To:   parentConfig,
	})
	if err != nil {
		log.Fatalln(err)
	}
	_, err = administrator.CreateInheritanceRel(context.TODO(), &oort.CreateInheritanceRelReq{
		From: childNamespace,
		To:   childConfig,
	})
	if err != nil {
		log.Fatalln(err)
	}
	_, err = administrator.CreateInheritanceRel(context.TODO(), &oort.CreateInheritanceRelReq{
		From: childNamespace,
		To:   app,
	})
	if err != nil {
		log.Fatalln(err)
	}
	// korisnik 1 moze da menja konfiguraciju unutar child ns-a
	_, err = administrator.CreatePolicy(context.TODO(), &oort.CreatePolicyReq{
		SubjectScope: user1,
		ObjectScope:  childNamespace,
		Permission:   putConfigPerm,
	})
	if err != nil {
		log.Fatalln(err)
	}

	// cekamo da se asinhrono poslati zahtevi obrade
	time.Sleep(2 * time.Second)

	// svi korisnici mogu da citaju konfiguraciju iz bilo kog ns-a
	// jer su nasledili dozvolu od grupe kojoj pripadaju
	// a ona je dodeljena nad roditeljskim ns-om -> vazi nad oba config resursa
	resp, err := evaluator.Authorize(context.TODO(), &oort.AuthorizationReq{
		Subject:        user1,
		Object:         parentConfig,
		PermissionName: getConfigPerm.Name,
	})
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(resp.Authorized)
	resp, err = evaluator.Authorize(context.TODO(), &oort.AuthorizationReq{
		Subject:        user1,
		Object:         childConfig,
		PermissionName: getConfigPerm.Name,
	})
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(resp.Authorized)
	resp, err = evaluator.Authorize(context.TODO(), &oort.AuthorizationReq{
		Subject:        user2,
		Object:         parentConfig,
		PermissionName: getConfigPerm.Name,
	})
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(resp.Authorized)
	resp, err = evaluator.Authorize(context.TODO(), &oort.AuthorizationReq{
		Subject:        user2,
		Object:         childConfig,
		PermissionName: getConfigPerm.Name,
	})
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(resp.Authorized)
	// korisnik jedan moze da menja sve konfiguracije unutar child ns
	// sve ostalo je zabranjeno
	resp, err = evaluator.Authorize(context.TODO(), &oort.AuthorizationReq{
		Subject:        user1,
		Object:         childNamespace,
		PermissionName: putConfigPerm.Name,
	})
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(resp.Authorized)
	resp, err = evaluator.Authorize(context.TODO(), &oort.AuthorizationReq{
		Subject:        user1,
		Object:         parentNamespace,
		PermissionName: putConfigPerm.Name,
	})
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(resp.Authorized)
	resp, err = evaluator.Authorize(context.TODO(), &oort.AuthorizationReq{
		Subject:        user2,
		Object:         childNamespace,
		PermissionName: putConfigPerm.Name,
	})
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(resp.Authorized)
	resp, err = evaluator.Authorize(context.TODO(), &oort.AuthorizationReq{
		Subject:        user2,
		Object:         parentNamespace,
		PermissionName: putConfigPerm.Name,
	})
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(resp.Authorized)
}

func grantedPermissions(evaluator oort.OortEvaluatorClient) {
	resp, err := evaluator.GetGrantedPermissions(context.TODO(), &oort.GetGrantedPermissionsReq{
		Subject: user1,
	})
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("permissions of user 1")
	for _, perm := range resp.Permissions {
		log.Printf("%s - %s/%s", perm.Name, perm.Object.Kind, perm.Object.Id)
	}

	resp, err = evaluator.GetGrantedPermissions(context.TODO(), &oort.GetGrantedPermissionsReq{
		Subject: user2,
	})
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("permissions of user 2")
	for _, perm := range resp.Permissions {
		log.Printf("%s - %s/%s", perm.Name, perm.Object.Kind, perm.Object.Id)
	}

	resp, err = evaluator.GetGrantedPermissions(context.TODO(), &oort.GetGrantedPermissionsReq{
		Subject: app,
	})
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("permissions of app 'my-app'")
	for _, perm := range resp.Permissions {
		log.Printf("%s - %s/%s", perm.Name, perm.Object.Kind, perm.Object.Id)
	}
}

var (
	parentNamespace = &oort.Resource{
		Id:   "parent",
		Kind: "namespace",
	}
	childNamespace = &oort.Resource{
		Id:   "child",
		Kind: "namespace",
	}
	parentConfig = &oort.Resource{
		Id:   "org1/config1/v1",
		Kind: "config",
	}
	childConfig = &oort.Resource{
		Id:   "org1/config1/v2",
		Kind: "config",
	}
	user1 = &oort.Resource{
		Id:   "1",
		Kind: "user",
	}
	user2 = &oort.Resource{
		Id:   "2",
		Kind: "user",
	}
	group = &oort.Resource{
		Id:   "1",
		Kind: "user-group",
	}
	app = &oort.Resource{
		Id:   "my-app",
		Kind: "app",
	}
	getConfigPerm = &oort.Permission{
		Name:      "config.get",
		Kind:      oort.Permission_ALLOW,
		Condition: &oort.Condition{Expression: ""},
	}
	denyGetConfigPerm = &oort.Permission{
		Name:      "config.get",
		Kind:      oort.Permission_DENY,
		Condition: &oort.Condition{Expression: ""},
	}
	putConfigPerm = &oort.Permission{
		Name:      "config.put",
		Kind:      oort.Permission_ALLOW,
		Condition: &oort.Condition{Expression: ""},
	}
)
