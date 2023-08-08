package main

import (
	"context"
	oort "github.com/c12s/oort/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

func main() {
	dial, err := grpc.Dial("localhost:8000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalln(err)
	}
	administrator := oort.NewOortAdministratorClient(dial)
	evaluator := oort.NewOortEvaluatorClient(dial)

	parentNamespace := &oort.Resource{
		Id:   "parent",
		Kind: "namespace",
	}
	childNamespace := &oort.Resource{
		Id:   "child",
		Kind: "namespace",
	}
	parentConfig := &oort.Resource{
		Id:   "parentConfig",
		Kind: "config",
	}
	childConfig := &oort.Resource{
		Id:   "childConfig",
		Kind: "config",
	}
	user1 := &oort.Resource{
		Id:   "1",
		Kind: "user",
	}
	user2 := &oort.Resource{
		Id:   "2",
		Kind: "user",
	}
	group := &oort.Resource{
		Id:   "1",
		Kind: "group",
	}
	getConfigPerm := &oort.Permission{
		Name:      "config.get",
		Kind:      oort.Permission_ALLOW,
		Condition: &oort.Condition{Expression: ""},
	}
	putConfigPerm := &oort.Permission{
		Name:      "config.put",
		Kind:      oort.Permission_ALLOW,
		Condition: &oort.Condition{Expression: ""},
	}

	// svi unutar grupe mogu da citaju konfiguracije unutar roditeljskog ns-a
	_, err = administrator.CreatePolicy(context.TODO(), &oort.CreatePolicyReq{
		SubjectScope: group,
		ObjectScope:  parentNamespace,
		Permission:   getConfigPerm,
	})
	if err != nil {
		log.Fatalln(err)
	}
	// korisnici nasledjuje dozvole iz grupe
	_, err = administrator.CreateInheritanceRel(context.TODO(), &oort.CreateInheritanceRelReq{
		From: group,
		To:   user1,
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
	// svaka konfiguracija pripada ns-u
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
	// korisnik 1 moze da menja konfiguraciju unutar child ns-a
	_, err = administrator.CreatePolicy(context.TODO(), &oort.CreatePolicyReq{
		SubjectScope: user1,
		ObjectScope:  childNamespace,
		Permission:   putConfigPerm,
	})
	if err != nil {
		log.Fatalln(err)
	}

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
	log.Println(resp.Allowed)
	resp, err = evaluator.Authorize(context.TODO(), &oort.AuthorizationReq{
		Subject:        user1,
		Object:         childConfig,
		PermissionName: getConfigPerm.Name,
	})
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(resp.Allowed)
	resp, err = evaluator.Authorize(context.TODO(), &oort.AuthorizationReq{
		Subject:        user2,
		Object:         childConfig,
		PermissionName: getConfigPerm.Name,
	})
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(resp.Allowed)
	resp, err = evaluator.Authorize(context.TODO(), &oort.AuthorizationReq{
		Subject:        user2,
		Object:         childConfig,
		PermissionName: getConfigPerm.Name,
	})
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(resp.Allowed)
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
	log.Println(resp.Allowed)
	resp, err = evaluator.Authorize(context.TODO(), &oort.AuthorizationReq{
		Subject:        user1,
		Object:         parentNamespace,
		PermissionName: putConfigPerm.Name,
	})
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(resp.Allowed)
	resp, err = evaluator.Authorize(context.TODO(), &oort.AuthorizationReq{
		Subject:        user2,
		Object:         childNamespace,
		PermissionName: putConfigPerm.Name,
	})
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(resp.Allowed)
	resp, err = evaluator.Authorize(context.TODO(), &oort.AuthorizationReq{
		Subject:        user2,
		Object:         parentNamespace,
		PermissionName: putConfigPerm.Name,
	})
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(resp.Allowed)
}
