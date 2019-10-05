
package main

import (
	"github.com/go-chi/chi"
	"net/http"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"strings"
	"github.com/DiGregory/tfs-auction/internal/auth"
	"strconv"
	"github.com/DiGregory/tfs-auction/internal/lot"
	"github.com/DiGregory/tfs-auction/internal/errors"
	"html/template"
	"github.com/gorilla/websocket"
	"log"
	"google.golang.org/grpc"
	"github.com/DiGregory/tfs-auction/cmd/lotpb"
	"context"
	"google.golang.org/grpc/status"
)

type RespError struct {
	Error string `json:"error"`
}

var UserLotsPageTmpl = template.Must(template.ParseFiles(
	`.\cmd\gateway-api\UserLots.html`,
))
var LotInfoPageTmpl = template.Must(template.ParseFiles(
	`.\cmd\gateway-api\LotInfo.html`,
))
var AllLotsPageTmpl = template.Must(template.ParseFiles(
	`.\cmd\gateway-api\AllLotsInfo.html`,
))

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func ShowError(w http.ResponseWriter, err error) {
	e := RespError{err.Error()}
	jsonResp, _ := json.Marshal(e)
	fmt.Println("Response: ", string(jsonResp))

	_,errWrite:=w.Write(jsonResp)
	if errWrite!=nil{
		fmt.Println("no data to response")
	}

}
func GetTokenFromHeader(r *http.Request) (*auth.Token) {
	t := r.Header.Get("Authorization")
	if t == "" {
		return nil
	}
	token := strings.Split(t, " ")
	return &auth.Token{TokenType: token[0], AccessToken: token[1]}
}




func (a *gateWayApp)CreateGatewayHandler() (error) {
	r := chi.NewRouter()

	r.Post("/signup", a.RegHandler)
	r.Post("/signin", a.SignInHandler)
	r.Put("/users/0", a.UpdUserHandler)
	r.Get("/users/{id}", a.GetUserHandler)

	r.Get("/users/{id}/lots", a.GetUserLotsHandler)
	r.Get("/users/{id}/getlots", a.HTMLUserLotsHandler)

	r.HandleFunc("/ws1", a.WSUpdateLot)
	r.Get("/lotget/{id}", a.GetLotWithWSHandler)

	r.HandleFunc("/ws2", a.WSUpdateAllLots)
	r.Get("/getlots", a.GetAllLotsWithWSHandler)

	r.Get("/lots", a.GetLotsHandler)
	r.Post("/lots", a.CreateLotHandler)
	r.Put("/lots/{id}", a.UpdateLotHandler)
	r.Get("/lots/{id}", a.GetSingleLot)
	r.Put("/lots/{id}/buy", a.BuyLotHandler)
	r.Delete("/lots/{id}", a.DeleteLotHandler)

	//бэкграунд процесс проверки старых лотов
	go a.LotStorage.CheckEndAt()

	fmt.Println("Server started at ", a.Addr)
	return http.ListenAndServe(a.Addr, r)

}
func (a *gateWayApp)BuyLotHandler(w http.ResponseWriter, r *http.Request) {

	ID := chi.URLParam(r, "id")
	LotID, err := strconv.Atoi(ID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		ShowError(w, errors.ErrNotFound)
		return
	}

	jsonReq, _ := ioutil.ReadAll(r.Body)

	fmt.Println("Buy Lot Request: \r\nLotID:", LotID, "\r\n", string(jsonReq))

	t := GetTokenFromHeader(r)
	UID, goodToken := auth.CheckToken(t)
	if !goodToken {
		w.WriteHeader(http.StatusUnauthorized)
		ShowError(w, errors.ErrUnAuth)
		return
	}
	conn, err := grpc.Dial("localhost:5001", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("can't connect to auction server: %v", err)
	}
	defer conn.Close()
	client := lotpb.NewLotsServiceClient(conn)
	req := lotpb.BuyLotRequest{
		JsonReq: jsonReq,
		LotID:   int64(LotID),
		UserID:  UID,
	}
	l, err := client.BuyLot(context.Background(), &req)

	statusCode, _ := status.FromError(err)

	switch {
	case err == nil:
		w.WriteHeader(http.StatusOK)
		jsonResp, _ := json.Marshal(l.Lot)
		fmt.Println("Response: ", string(jsonResp))
		_,errWrite:=w.Write(jsonResp)
		if errWrite!=nil{
			fmt.Println("no data to response")
		}
	case errors.ErrBadReq.Error() == statusCode.Message():
		w.WriteHeader(http.StatusBadRequest)
		ShowError(w, errors.ErrBadReq)
	case errors.ErrNotFound.Error() == statusCode.Message():
		w.WriteHeader(http.StatusNotFound)
		ShowError(w, errors.ErrNotFound)
	case errors.ErrConflict.Error() == statusCode.Message():
		w.WriteHeader(http.StatusConflict)
		ShowError(w, errors.ErrConflict)
	}

}

func (a *gateWayApp)GetSingleLot(w http.ResponseWriter, r *http.Request) {
	ID := chi.URLParam(r, "id")
	LotID, err := strconv.Atoi(ID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		ShowError(w, errors.ErrNotFound)
		return
	}
	fmt.Println("Get  Lot Request: \r\nLotID:", LotID)

	t := GetTokenFromHeader(r)
	_, goodToken := auth.CheckToken(t)
	if !goodToken {
		w.WriteHeader(http.StatusUnauthorized)
		ShowError(w, errors.ErrUnAuth)
		return
	}

	conn, err := grpc.Dial("localhost:5001", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("can't connect to auction server: %v", err)
	}
	defer conn.Close()
	client := lotpb.NewLotsServiceClient(conn)
	req := lotpb.SingleLotRequestId{
		Id: int64(LotID),
	}
	l, err := client.GetSingleLot(context.Background(), &req)

	statusCode, _ := status.FromError(err)

	switch {
	case err == nil:
		w.WriteHeader(http.StatusOK)
		jsonResp, _ := json.Marshal(l.Lot)
		fmt.Println("Response: ", string(jsonResp))
		_,errWrite:=w.Write(jsonResp)
		if errWrite!=nil{
			fmt.Println("no data to response")
		}
	case errors.ErrBadReq.Error() == statusCode.Message():
		w.WriteHeader(http.StatusBadRequest)
		ShowError(w, errors.ErrBadReq)

	case errors.ErrNotFound.Error() == statusCode.Message():
		w.WriteHeader(http.StatusNotFound)
		ShowError(w, errors.ErrNotFound)

	}

}

func (a *gateWayApp)GetLotsHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Get All Lots Request:")
	Status := ""
	if _, ok := r.URL.Query()["status"]; ok {
		Status = r.URL.Query()["status"][0]
	}

	if Status != "" && Status != "created" && Status != "active" && Status != "finished" {
		w.WriteHeader(http.StatusBadRequest)
		ShowError(w, errors.ErrBadReq)
		return
	}
	t := GetTokenFromHeader(r)
	_, goodToken := auth.CheckToken(t)
	if !goodToken {
		w.WriteHeader(http.StatusUnauthorized)
		ShowError(w, errors.ErrUnAuth)
		return
	}
	conn, err := grpc.Dial("localhost:5001", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("can't connect to auction server: %v", err)
	}
	defer conn.Close()

	client := lotpb.NewLotsServiceClient(conn)
	req := lotpb.AllLotsRequestStatus{
		Status: Status,
	}
	l, err := client.GetAllLots(context.Background(), &req)
	statusCode, _ := status.FromError(err)

	switch {
	case err == nil:
		w.WriteHeader(http.StatusOK)
		jsonResp, _ := json.Marshal(l.Lots)
		fmt.Println("Response: ", string(jsonResp))
		_,errWrite:=w.Write(jsonResp)
		if errWrite!=nil{
			fmt.Println("no data to response")
		}

	case errors.ErrBadReq.Error() == statusCode.Message():
		w.WriteHeader(http.StatusBadRequest)
		ShowError(w, errors.ErrBadReq)
	case errors.ErrNotFound.Error() == statusCode.Message():
		w.WriteHeader(http.StatusNotFound)
		ShowError(w, errors.ErrNotFound)

	}

}

func (a *gateWayApp)GetLotWithWSHandler(w http.ResponseWriter, r *http.Request) {
	ID := chi.URLParam(r, "id")
	LotID, err := strconv.Atoi(ID)

	fmt.Println("Request UpdateLot with WS: \r\n LotID:", LotID)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		ShowError(w, errors.ErrNotFound)
		return
	}

	t := GetTokenFromHeader(r)
	_, goodToken := auth.CheckToken(t)
	if !goodToken {
		w.WriteHeader(http.StatusUnauthorized)
		ShowError(w, errors.ErrUnAuth)
		return
	}

	l, err := a.LotStorage.GetLot(LotID)

	switch err {
	case nil:
		w.WriteHeader(http.StatusOK)
		var data = map[string]interface{}{}
		var err error
		data["lot"] = l

		err = LotInfoPageTmpl.ExecuteTemplate(w, "LotInfo", data)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			ShowError(w, errors.ErrBadReq)
			return
		}

	case errors.ErrNotFound:
		w.WriteHeader(http.StatusNotFound)
		ShowError(w, errors.ErrNotFound)

	case errors.ErrBadReq:
		w.WriteHeader(http.StatusBadRequest)
		ShowError(w, errors.ErrBadReq)

	}

}

func (a *gateWayApp)GetAllLotsWithWSHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get All Lots Request with WS:")
	Status := ""
	if _, ok := r.URL.Query()["status"]; ok {
		Status = r.URL.Query()["status"][0]
	}

	if Status != "" && Status != "created" && Status != "active" && Status != "finished" {
		w.WriteHeader(http.StatusBadRequest)
		ShowError(w, errors.ErrBadReq)
		return
	}
	t := GetTokenFromHeader(r)
	_, goodToken := auth.CheckToken(t)
	if !goodToken {
		w.WriteHeader(http.StatusUnauthorized)
		ShowError(w, errors.ErrUnAuth)
		return
	}

	l, err := a.LotStorage.GetAllLots(Status)

	switch err {
	case nil:
		w.WriteHeader(http.StatusOK)
		var data = map[string]interface{}{}
		var err error
		data["lots"] = l

		err = AllLotsPageTmpl.ExecuteTemplate(w, "AllLotsInfo", data)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			ShowError(w, errors.ErrBadReq)
			return
		}
	case errors.ErrBadReq:
		w.WriteHeader(http.StatusBadRequest)
		ShowError(w, errors.ErrBadReq)
	case errors.ErrNotFound:
		w.WriteHeader(http.StatusNotFound)
		ShowError(w, errors.ErrNotFound)

	}
}

func (a *gateWayApp)WSUpdateAllLots(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("can't upgrade connection: %s\n", err)
		return
	}
	defer conn.Close()

	lot.WsClients2.AddClient(conn)

	for {
		AllLots := <-lot.AllLotsCh
		res, err := json.Marshal(AllLots)
		if err != nil {
			fmt.Printf("can't marshal message: %+v\n", err)
			continue
		}
		lot.WsClients2.BroadcastMessage(res)
	}

}

func (a *gateWayApp)WSUpdateLot(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("can't upgrade connection: %s\n", err)
		return
	}
	defer conn.Close()
	lot.WsClients1.AddClient(conn)

	for {
		SingleLot := <-lot.SingleLotCh
		res, err := json.Marshal(SingleLot)
		if err != nil {
			fmt.Printf("can't marshal message: %+v\n", err)
			continue
		}
		lot.WsClients1.BroadcastMessage(res)
	}

}

func (a *gateWayApp)CreateLotHandler(w http.ResponseWriter, r *http.Request) {
	jsonReq, _ := ioutil.ReadAll(r.Body)
	fmt.Println("Create Lot Request: ", string(jsonReq))
	t := GetTokenFromHeader(r)
	UID, goodToken := auth.CheckToken(t)
	if !goodToken {
		w.WriteHeader(http.StatusUnauthorized)
		ShowError(w, errors.ErrUnAuth)
		return
	}

	conn, err := grpc.Dial("localhost:5001", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("can't connect to auction server: %v", err)
	}
	defer conn.Close()

	client := lotpb.NewLotsServiceClient(conn)
	req := lotpb.CreateLotRequest{
		JsonReq: jsonReq,
		UserID:  UID,
	}
	l, err := client.CreateLot(context.Background(), &req)

	statusCode, _ := status.FromError(err)



	switch   {
	case err==nil:
		jsonResp, _ := json.Marshal(l.Lot)
		fmt.Println("Response: ", string(jsonResp))
		_,errWrite:=w.Write(jsonResp)
		if errWrite!=nil{
			fmt.Println("no data to response")
		}
	case errors.ErrBadReq.Error()==statusCode.Message():
		w.WriteHeader(http.StatusBadRequest)
		ShowError(w, errors.ErrBadReq)

	}
}

func (a *gateWayApp)GetUserLotsHandler(w http.ResponseWriter, r *http.Request) {
	ID := chi.URLParam(r, "id")
	UserID, err := strconv.Atoi(ID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		ShowError(w, errors.ErrNotFound)
		return
	}
	fmt.Println("Get User Lots Request: \r\nId:", UserID)
	Type := ""
	if _, ok := r.URL.Query()["type"]; ok {
		Type = r.URL.Query()["type"][0]
	}

	if Type != "" && Type != "own" && Type != "buyed" {
		w.WriteHeader(http.StatusBadRequest)
		ShowError(w, errors.ErrBadReq)
		return
	}
	t := GetTokenFromHeader(r)
	UID, goodToken := auth.CheckToken(t)
	if !goodToken {
		w.WriteHeader(http.StatusUnauthorized)
		ShowError(w, errors.ErrUnAuth)
		return
	}

	if UserID == 0 {
		UserID = int(UID)
	}



	conn, err := grpc.Dial("localhost:5001", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("can't connect to auction server: %v", err)
	}
	defer conn.Close()

	client := lotpb.NewLotsServiceClient(conn)
	req := lotpb.AllUserLotsRequest{
		Type: Type,

		UserID:  UID,
	}
	l, err := client.GetAllUserLots(context.Background(), &req)

	statusCode, _ := status.FromError(err)


	switch   {
	case err==nil:
		w.WriteHeader(http.StatusOK)
		jsonResp, _ := json.Marshal(l.Lots)
		fmt.Println("Response: ", string(jsonResp))
		_,errWrite:=w.Write(jsonResp)
		if errWrite!=nil{
			fmt.Println("no data to response")
		}
	case errors.ErrBadReq.Error()==statusCode.Message():
		w.WriteHeader(http.StatusBadRequest)
		ShowError(w, errors.ErrBadReq)

	case errors.ErrNotFound.Error()==statusCode.Message():
		w.WriteHeader(http.StatusNotFound)
		ShowError(w, errors.ErrNotFound)

	}

}

func (a *gateWayApp)HTMLUserLotsHandler(w http.ResponseWriter, r *http.Request) {
	ID := chi.URLParam(r, "id")
	UserID, err := strconv.Atoi(ID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		ShowError(w, errors.ErrNotFound)
		return
	}
	fmt.Println("HTML User Get Lots Request: \r\nId:", UserID)
	Type := ""
	if _, ok := r.URL.Query()["type"]; ok {
		Type = r.URL.Query()["type"][0]
	}

	if Type != "" && Type != "own" && Type != "buyed" {
		w.WriteHeader(http.StatusBadRequest)
		ShowError(w, errors.ErrBadReq)
		return
	}
	t := GetTokenFromHeader(r)
	UID, goodToken := auth.CheckToken(t)
	if !goodToken {
		w.WriteHeader(http.StatusUnauthorized)
		ShowError(w, errors.ErrUnAuth)
		return
	}

	if UserID == 0 {
		UserID = int(UID)
	}

	conn, err := grpc.Dial("localhost:5001", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("can't connect to auction server: %v", err)
	}
	defer conn.Close()

	client := lotpb.NewLotsServiceClient(conn)
	req := lotpb.AllUserLotsRequest{
		Type: Type,

		UserID:  UID,
	}
	l, err := client.GetAllUserLots(context.Background(), &req)

	statusCode, _ := status.FromError(err)

	switch   {
	case err==nil:
		w.WriteHeader(http.StatusOK)
		var data = map[string]interface{}{}
		var err error
		data["lots"] = l.Lots

		err = UserLotsPageTmpl.ExecuteTemplate(w, "UserLots", data)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			ShowError(w, errors.ErrBadReq)
			return
		}

	case errors.ErrBadReq.Error()==statusCode.Message():
		w.WriteHeader(http.StatusBadRequest)
		ShowError(w, errors.ErrBadReq)

	case errors.ErrNotFound.Error()==statusCode.Message():
		w.WriteHeader(http.StatusNotFound)
		ShowError(w, errors.ErrNotFound)

	}

}

func (a *gateWayApp)UpdateLotHandler(w http.ResponseWriter, r *http.Request) {
	ID := chi.URLParam(r, "id")
	LotID, err := strconv.Atoi(ID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		ShowError(w, errors.ErrNotFound)
		return
	}

	jsonReq, _ := ioutil.ReadAll(r.Body)
	fmt.Println("Update LotRequest: \r\nId:", LotID, "\r\n BodyRequest: ", string(jsonReq))

	t := GetTokenFromHeader(r)
	UID, goodToken := auth.CheckToken(t)
	if !goodToken {
		w.WriteHeader(http.StatusUnauthorized)
		ShowError(w, errors.ErrUnAuth)
		return
	}

	conn, err := grpc.Dial("localhost:5001", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("can't connect to auction server: %v", err)
	}
	defer conn.Close()

	client := lotpb.NewLotsServiceClient(conn)
	req := lotpb.UpdateLotRequest{
		JsonReq: jsonReq,
		LotID:   int64(LotID),
		UserID:  UID,
	}
	l, err := client.UpdateLot(context.Background(), &req)

	statusCode, _ := status.FromError(err)

	switch {
	case err == nil:
		w.WriteHeader(http.StatusOK)
		jsonResp, _ := json.Marshal(l.Lot)
		fmt.Println("Response: ", string(jsonResp))
		_,errWrite:=w.Write(jsonResp)
		if errWrite!=nil{
			fmt.Println("no data to response")
		}
	case errors.ErrBadReq.Error() == statusCode.Message():
		w.WriteHeader(http.StatusBadRequest)
		ShowError(w, errors.ErrBadReq)

	case errors.ErrNotFound.Error() == statusCode.Message():
		w.WriteHeader(http.StatusNotFound)
		ShowError(w, errors.ErrNotFound)

	}
}

func (a *gateWayApp)DeleteLotHandler(w http.ResponseWriter, r *http.Request) {
	ID := chi.URLParam(r, "id")
	LotID, err := strconv.Atoi(ID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		ShowError(w, errors.ErrNotFound)
		return
	}

	fmt.Println("Delete Lot Request: \r\nId:", LotID)

	t := GetTokenFromHeader(r)
	UID, goodToken := auth.CheckToken(t)
	if !goodToken {
		w.WriteHeader(http.StatusUnauthorized)
		ShowError(w, errors.ErrUnAuth)
		return
	}

	conn, err := grpc.Dial("localhost:5001", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("can't connect to auction server: %v", err)
	}
	defer conn.Close()

	client := lotpb.NewLotsServiceClient(conn)
	req := lotpb.DeleteLotRequest{
		LotID:  int64(LotID),
		UserID: UID,
	}
	_, err = client.DeleteLot(context.Background(), &req)

	statusCode, _ := status.FromError(err)

	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)

	case errors.ErrNotFound.Error() == statusCode.Message():
		w.WriteHeader(http.StatusNotFound)

	}
}

func (a *gateWayApp)SignInHandler(w http.ResponseWriter, r *http.Request) {
	jsonReq, _ := ioutil.ReadAll(r.Body)
	fmt.Println("Sign In Request: ", string(jsonReq))
	Token, err := auth.CheckAuth(jsonReq)

	switch err {
	case nil:
		w.WriteHeader(http.StatusOK)
		jsonResp, _ := json.Marshal(Token)
		fmt.Println("Response: ", string(jsonResp))
		_,errWrite:=w.Write(jsonResp)
		if errWrite!=nil{
			fmt.Println("no data to response")
		}

	case errors.ErrInvalidEmailOrPass:
		w.WriteHeader(http.StatusUnauthorized)
		ShowError(w, errors.ErrInvalidEmailOrPass)

	}

}

func (a *gateWayApp)GetUserHandler(w http.ResponseWriter, r *http.Request) {
	ID := r.URL.Path[len("/users/"):]
	UserID, err := strconv.Atoi(ID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		ShowError(w, errors.ErrNotFound)
		return
	}

	fmt.Println("Get User Request: ", UserID)

	t := GetTokenFromHeader(r)
	UID, goodToken := auth.CheckToken(t)
	if !goodToken {
		w.WriteHeader(http.StatusUnauthorized)
		ShowError(w, errors.ErrUnAuth)
		return
	}

	if UserID == 0 {
		UserID = int(UID)
	}
	u, err := a.UserStorage.GetUser(UserID)

	switch err {
	case nil:
		w.WriteHeader(http.StatusOK)
		jsonResp, _ := json.Marshal(u)
		fmt.Println("Response: ", string(jsonResp))
		_,errWrite:=w.Write(jsonResp)
		if errWrite!=nil{
			fmt.Println("no data to response")
		}
	case errors.ErrNotFound:
		w.WriteHeader(http.StatusNotFound)
		ShowError(w, errors.ErrNotFound)

	}
}

func (a *gateWayApp)RegHandler(w http.ResponseWriter, r *http.Request) {

	jsonReq, err := ioutil.ReadAll(r.Body)
	if err!=nil{
		w.WriteHeader(http.StatusBadRequest)
		ShowError(w, errors.ErrBadReq)
	}
	fmt.Println("Reg User Request: ", string(jsonReq))
	err = a.UserStorage.RegUser(jsonReq)

	switch err {
	case nil:
		w.WriteHeader(http.StatusCreated)
	case errors.ErrBadReq:
		w.WriteHeader(http.StatusBadRequest)
		ShowError(w, errors.ErrBadReq)

	case errors.ErrEmailExist:
		w.WriteHeader(http.StatusConflict)
		ShowError(w, errors.ErrEmailExist)
	}

}

func (a *gateWayApp)UpdUserHandler(w http.ResponseWriter, r *http.Request) {
	jsonReq, _ := ioutil.ReadAll(r.Body)
	fmt.Println("Request: ", string(jsonReq))
	t := GetTokenFromHeader(r)
	UID, goodToken := auth.CheckToken(t)
	if !goodToken {
		w.WriteHeader(http.StatusUnauthorized)
		ShowError(w, errors.ErrUnAuth)
		return
	}

	u, err := a.UserStorage.UpdateUser(jsonReq, UID)

	switch err {
	case nil:
		w.WriteHeader(http.StatusOK)
		jsonResp, _ := json.Marshal(u)
		fmt.Println("Response: ", string(jsonResp))
		_,errWrite:=w.Write(jsonResp)
		if errWrite!=nil{
			fmt.Println("no data to response")
		}
	case errors.ErrBadReq:
		w.WriteHeader(http.StatusBadRequest)
		ShowError(w, errors.ErrBadReq)

	}
}
