package main

import (
	"context"
	"database/sql"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	pb "grpc_service/proto"
	"log"
	"net"
)

type server struct {
	pb.UnimplementedZoneServiceServer
	db *sql.DB
}

func (s *server) GetZones(ctx context.Context, _ *pb.Empty) (*pb.ZoneList, error) {
	rows, err := s.db.Query("select id, name, zone_factor from zone")

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch")
	}
	defer rows.Close()
	var zones []*pb.ZoneRes

	for rows.Next() {
		var z pb.ZoneRes
		var name sql.NullString
		var factor sql.NullFloat64
		err := rows.Scan(&z.Id, &name, &factor)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if name.Valid {
			z.Name = name.String
		} else {
			z.Name = ""
		}

		if factor.Valid {
			z.ZoneFactor = factor.Float64
		} else {
			z.ZoneFactor = 0
		}

		zones = append(zones, &z)

	}
	return &pb.ZoneList{Zones: zones}, nil

}

func (s *server) GetZoneByID(ctx context.Context, req *pb.ZoneReq) (*pb.ZoneRes, error) {
	var z pb.ZoneRes
	var name sql.NullString
	var factor sql.NullFloat64
	err := s.db.QueryRow("select id, name, zone_factor from zone where id=$1", req.Id).Scan(&z.Id, name, factor)

	if err == sql.ErrNoRows {
		return nil, status.Error(codes.NotFound, "zone not found")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &z, nil
}

func (s *server) CreateZone(ctx context.Context, req *pb.ZoneData) (*pb.ZoneRes, error) {
	var z pb.ZoneRes
	var id int32
	err := s.db.QueryRow("Insert into zone(name, code_name, zone_factor) values ($1, $2, $3) RETURNING id", req.Name, req.CodeName, req.ZoneFactor).Scan(&id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	z.Id = id
	z.Name = req.Name
	z.CodeName = req.CodeName
	z.ZoneFactor = req.ZoneFactor

	return &z, nil
}

func( s *server) UpdateZone (ctx context.Context, req *pb.UpdateZoneReq) (*pb.ZoneRes,error){
	res, err:= s.db.Exec("UPDATE zone SET name=$1, code_name=$2, zone_factor=$3 WHERE id=$4", req.Name, req.CodeName, req.ZoneFactor, req.Id )

	if err!=nil{
		return nil, status.Error(codes.Internal, err.Error())
	}
	rows, _:= res.RowsAffected()
	if rows==0{
		return nil, status.Error(codes.NotFound, "zone not found")
	}
	return &pb.ZoneRes{Id: req.Id, Name: req.Name, CodeName: req.CodeName, ZoneFactor: req.ZoneFactor}, nil
}

func (s *server) DeleteZone(ctx context.Context, req *pb.ZoneReq) (*pb.DeleteRes, error){
	res, err:= s.db.Exec("DELETE FROM zone WHERE id=$1",req.Id)
	if err!=nil {
		return nil, status.Error(codes.NotFound, "zone not found")
	}
	rows,_:=res.RowsAffected()
	if rows==0{
		return nil, status.Error(codes.NotFound, "zone not found")
	}
	return &pb.DeleteRes{Message: "Koppp"}, nil
}

func main() {
	connStr := "host=localhost port=5432 user=rakib password=123 dbname=test_map_eta sslmode=disable"

	db, err := sql.Open("postgres", connStr)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal("not conneted", err)
	}
	log.Println("connected to database")
	lis, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterZoneServiceServer(grpcServer, &server{db: db})
	reflection.Register(grpcServer)
	log.Print("grpc server is on 8081")
	log.Fatal(grpcServer.Serve(lis))

}
