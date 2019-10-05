package main

import (
	"net"
	"log"
	"google.golang.org/grpc"
	"github.com/DiGregory/tfs-auction/cmd/lotpb"
	"context"
	"github.com/DiGregory/tfs-auction/internal/lot"
	"fmt"
)

var DSN="user=postgres password=1234 dbname=test sslmode=disable"

type server struct{}
func (*server)CreateLot(ctx context.Context, req *lotpb.CreateLotRequest)(*lotpb.CreatedLotResponse, error) {
	s,err:=lot.StorageConnect(DSN)
	if err!=nil{
		log.Fatal(err)
	}
	createdLot,err:=s.CreateLot(req.JsonReq,req.UserID)
	if err != nil {
		return nil, err
	}

	resp := lotpb.CreatedLotResponse{
		 Lot:&lotpb.Lot{
			ID:          int64(createdLot.ID),
			Title:       createdLot.Title,
			Description: createdLot.Description,
			BuyPrice:    createdLot.BuyPrice,
			MinPrice:    createdLot.MinPrice,
			PriceStep:   createdLot.PriceStep,
			Status:      createdLot.Status,
			EndAt:       createdLot.EndAt.String(),
			CreatedAt:   createdLot.CreatedAt.String(),
			UpdateAt:    createdLot.UpdatedAt.String(),
			Creator:     &lotpb.ShortUser{ID: createdLot.Creator.ID, FirstName: createdLot.Creator.FirstName, LastName: createdLot.Creator.LastName},
		},
	}
	return &resp,nil
}
func (*server) DeleteLot(ctx context.Context, req *lotpb.DeleteLotRequest) (*lotpb.Nothing, error) {
	s,err:=lot.StorageConnect(DSN)
	if err!=nil{
		log.Fatal(err)
	}
	err  = s.DeleteLot(req.UserID, uint(req.LotID))
	if err != nil {
		return nil, err
	}

	return &lotpb.Nothing{Dummy: true}, nil
}
func (*server) BuyLot(ctx context.Context, req *lotpb.BuyLotRequest) (*lotpb.BuyingLotResponse, error) {
	s,err:=lot.StorageConnect(DSN)
	if err!=nil{
		log.Fatal(err)
	}
	BuyedLot, err := s.BuyLot(req.JsonReq, int(req.LotID), req.UserID)
	if err != nil {
		return nil, err
	}

	resp := lotpb.BuyingLotResponse{
		Lot: &lotpb.Lot{
			ID:          int64(BuyedLot.ID),
			Title:       BuyedLot.Title,
			Description: BuyedLot.Description,
			BuyPrice:    BuyedLot.BuyPrice,
			MinPrice:    BuyedLot.MinPrice,
			PriceStep:   BuyedLot.PriceStep,
			Status:      BuyedLot.Status,
			EndAt:       BuyedLot.EndAt.String(),
			CreatedAt:   BuyedLot.CreatedAt.String(),
			UpdateAt:    BuyedLot.UpdatedAt.String(),
			Creator:     &lotpb.ShortUser{ID: BuyedLot.Creator.ID, FirstName: BuyedLot.Creator.FirstName, LastName: BuyedLot.Creator.LastName},
		},
	}
	if BuyedLot.Buyer != nil {
		resp.Lot.Buyer = &lotpb.ShortUser{ID: BuyedLot.Buyer.ID, FirstName: BuyedLot.Buyer.FirstName, LastName: BuyedLot.Buyer.LastName}
	}
	return &resp, nil
}
func (*server) UpdateLot(ctx context.Context, req *lotpb.UpdateLotRequest) (*lotpb.UpdatedLotResponse, error) {
	s,err:=lot.StorageConnect(DSN)
	if err!=nil{
		log.Fatal(err)
	}
	updatedLot, err := s.UpdateLot(req.JsonReq, req.UserID, int(req.LotID))
	if err != nil {
		return nil, err
	}


	resp := lotpb.UpdatedLotResponse{
		Lot: &lotpb.Lot{
			ID:          int64(updatedLot.ID),
			Title:       updatedLot.Title,
			Description: updatedLot.Description,
			BuyPrice:    updatedLot.BuyPrice,
			MinPrice:    updatedLot.MinPrice,
			PriceStep:   updatedLot.PriceStep,
			Status:      updatedLot.Status,
			EndAt:       updatedLot.EndAt.String(),
			CreatedAt:   updatedLot.CreatedAt.String(),
			UpdateAt:    updatedLot.UpdatedAt.String(),
			Creator:     &lotpb.ShortUser{ID: updatedLot.Creator.ID, FirstName: updatedLot.Creator.FirstName, LastName: updatedLot.Creator.LastName},
		},
	}
	fmt.Println(resp)
	if updatedLot.Buyer != nil {
		resp.Lot.Buyer = &lotpb.ShortUser{ID: updatedLot.Buyer.ID, FirstName: updatedLot.Buyer.FirstName, LastName: updatedLot.Buyer.LastName}
	}
	return &resp, nil
}
func(*server)GetAllUserLots(ctx context.Context, req *lotpb.AllUserLotsRequest) (*lotpb.AllUserLotsResponse, error){
	s,err:=lot.StorageConnect(DSN)
	if err!=nil{
		log.Fatal(err)
	}
	lots,err:=s.GetUserLots(int(req.UserID),req.Type)
	if err != nil {
		return nil, err
	}
	AllLots := lotpb.AllUserLotsResponse{}

	if len(lots) != 0 {

		for i := range lots {
			newLot := &lotpb.Lot{
				ID:          int64(lots[i].ID),
				Title:       lots[i].Title,
				Description: lots[i].Description,
				BuyPrice:    lots[i].BuyPrice,
				MinPrice:    lots[i].MinPrice,
				PriceStep:   lots[i].PriceStep,
				Status:      lots[i].Status,
				EndAt:       lots[i].EndAt.String(),
				CreatedAt:   lots[i].CreatedAt.String(),
				UpdateAt:    lots[i].UpdatedAt.String(),
				Creator:     &lotpb.ShortUser{ID: lots[i].Creator.ID, FirstName: lots[i].Creator.FirstName, LastName: lots[i].Creator.LastName},
			}
			if lots[i].Buyer != nil {
				newLot.Buyer = &lotpb.ShortUser{ID: lots[i].Creator.ID, FirstName: lots[i].Creator.FirstName, LastName: lots[i].Creator.LastName}
			}
			AllLots.Lots = append(AllLots.Lots, newLot)
		}
	}

	return &AllLots, nil
}
func (*server) GetAllLots(ctx context.Context, req *lotpb.AllLotsRequestStatus) (*lotpb.AllLotsResponse, error) {
	s,err:=lot.StorageConnect(DSN)
	if err!=nil{
		log.Fatal(err)
	}
	l, err := s.GetAllLots(req.Status)
	if err != nil {
		return nil, err
	}
	AllLots := lotpb.AllLotsResponse{}

	if len(l) != 0 {

		for i := range l {
			newLot := &lotpb.Lot{
				ID:          int64(l[i].ID),
				Title:       l[i].Title,
				Description: l[i].Description,
				BuyPrice:    l[i].BuyPrice,
				MinPrice:    l[i].MinPrice,
				PriceStep:   l[i].PriceStep,
				Status:      l[i].Status,
				EndAt:       l[i].EndAt.String(),
				CreatedAt:   l[i].CreatedAt.String(),
				UpdateAt:    l[i].UpdatedAt.String(),
				Creator:     &lotpb.ShortUser{ID: l[i].Creator.ID, FirstName: l[i].Creator.FirstName, LastName: l[i].Creator.LastName},
			}
			if l[i].Buyer != nil {
				newLot.Buyer = &lotpb.ShortUser{ID: l[i].Creator.ID, FirstName: l[i].Creator.FirstName, LastName: l[i].Creator.LastName}
			}
			AllLots.Lots = append(AllLots.Lots, newLot)
		}
	}

	return &AllLots, nil
}

func (*server) GetSingleLot(ctx context.Context, req *lotpb.SingleLotRequestId) (*lotpb.SingleLotResponse, error) {
	s,err:=lot.StorageConnect(DSN)
	if err!=nil{
		log.Fatal(err)
	}
	lot, err := s.GetLot(int(req.Id))

	if err != nil {
		return nil, err
	}

	resp := lotpb.SingleLotResponse{
		Lot: &lotpb.Lot{
			ID:          int64(lot.ID),
			Title:       lot.Title,
			Description: lot.Description,
			BuyPrice:    lot.BuyPrice,
			MinPrice:    lot.MinPrice,
			PriceStep:   lot.PriceStep,
			Status:      lot.Status,
			EndAt:       lot.EndAt.String(),
			CreatedAt:   lot.CreatedAt.String(),
			UpdateAt:    lot.UpdatedAt.String(),
			Creator:     &lotpb.ShortUser{ID: lot.Creator.ID, FirstName: lot.Creator.FirstName, LastName: lot.Creator.LastName},
		},
	}
	if lot.Buyer != nil {
		resp.Lot.Buyer = &lotpb.ShortUser{ID: lot.Buyer.ID, FirstName: lot.Buyer.FirstName, LastName: lot.Buyer.LastName}
	}
	return &resp, nil
}

func main() {

	network:="tcp"
	addr:=":5001"

	listen, err := net.Listen(network, addr)
	if err != nil {
		log.Fatalf("can't listen on port: %v", err)
	}

	s := grpc.NewServer()
	lotpb.RegisterLotsServiceServer(s, &server{})
	if err := s.Serve(listen); err != nil {
		log.Fatalf("can't register service server: %v", err)
	}

}
