package main

import (
	"calculator/calculatorpb"
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

type server struct{}

func (*server) Sum(ctx context.Context,req *calculatorpb.SumRequest) (*calculatorpb.SumResponse, error){
	log.Println("Sum call")
	resp := &calculatorpb.SumResponse{
		Result: req.GetNum1() + req.GetNum2(),
	}
	return resp, nil
}
func (*server) SumWithDeadLine(ctx context.Context,req *calculatorpb.SumRequest) (*calculatorpb.SumResponse, error){
	log.Println("Sum with dead line call")
	for i := 0; i < 3; i++ {
		if ctx.Err() == context.Canceled {
			log.Println("context.Canceled...")
			return nil, status.Error(codes.Canceled, " client canceled req")
		}
		time.Sleep(1000*time.Millisecond)
	}
	resp := &calculatorpb.SumResponse{
		Result: req.GetNum1() + req.GetNum2(),
	}
	return resp, nil
}
func (*server) PrimeNumber(req *calculatorpb.PNRequest,
	stream calculatorpb.CalculatorService_PrimeNumberServer) error{
		log.Println("Prime Number call")
		k := int32(2)
		N := req.GetNumber()
		for N > 1 {
			if N%k == 0 {
				N = N / k
				//send to client
				stream.Send(&calculatorpb.PNResponse{
					Result : k,
				})
				time.Sleep(500*time.Millisecond)
			} else {
				k++
				log.Printf("k inc %v",k)
			}
		}
		return nil
}
func (*server) Average(stream calculatorpb.CalculatorService_AverageServer) error{
	log.Println("AVG called")
	var total float32
	var count int
	for {
		req, err := stream.Recv()
		if err == io.EOF{
			resp := &calculatorpb.AVGResponse{
				Result: total / float32(count),
			}
			return stream.SendAndClose(resp)
		}
		if err != nil {
			log.Fatalf("err while Recv AVG %v", err)
			return err
		}
		log.Printf("recv num %v", req)
		total += req.GetNumber()
		count++
	}
}
func(*server) FindMax(stream calculatorpb.CalculatorService_FindMaxServer) error{
	log.Println("Find max called")
	max := int32(0)
	for {
		req, err := stream.Recv()
		if err == io.EOF{
			log.Println("EOF...")
			return nil
		}
		if err != nil {
			log.Fatalf("err while Recv FindMax %v", err)
			return err
		}
		log.Printf("recv num %v", req)
		num := req.GetNum()
		if num > max {
			max = num
		}
		err = stream.Send(&calculatorpb.FMResponse{
			Max: max,
		})
		if err != nil {
			log.Fatalf("send maxx err %v", err)
			return err
		}
	}
}
func (*server) Square(ctx context.Context,req *calculatorpb.SquareRequest) (*calculatorpb.SquareResponse, error){
	log.Println("Square called")
	num := req.GetNum()
	if num < 0 {
		log.Printf(" req num < 0, num=%v, return InvalArgument", num)
		return nil, status.Errorf(codes.InvalidArgument,"Expect num > 0, req num was %v", num)
	}
	return &calculatorpb.SquareResponse{
		Result: math.Sqrt(float64(num)),
	}, nil
}
func main(){

	lis, err := net.Listen("tcp","0.0.0.0:5000")
	if err != nil {
		log.Fatal("err create listen %w",err)
	}
	certFile := "ssl/server.crt"
	keyFile := "ssl/server.pem"
	creds, sslErr := credentials.NewServerTLSFromFile(certFile, keyFile)
	if sslErr != nil {
		log.Fatalf("create creds ssl err %v\n", sslErr)
		return
	}
	opt := grpc.Creds(creds)
	s := grpc.NewServer(opt)
	calculatorpb.RegisterCalculatorServiceServer(s, &server{})

	fmt.Println("calculator is running ...")
	err = s.Serve(lis)
	if err != nil {
		log.Fatal("errr %w", err)
	}
}
