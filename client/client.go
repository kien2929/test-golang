package main

import (
	"calculator/calculatorpb"
	"context"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

func main() {
	certFile := "ssl/ca.crt"
	creds, sslErr := credentials.NewClientTLSFromFile(certFile, "")
	if sslErr != nil{
		log.Fatalf(" create client creds ssl err %v\n", sslErr)
	}

	cc, err := grpc.Dial("localhost:5000", grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("err while dial %v", err)
	}
	defer cc.Close()
	client := calculatorpb.NewCalculatorServiceClient(cc)
	callSum(client)
	// callPND(client)
	// callAVG(client)
	// callFindMax(client)
	// callSquare(client, -4)
	// callSumWithDeadLine(client, 1 * time.Second)
	// callSumWithDeadLine(client, 5 * time.Second)
}

func callSum(c calculatorpb.CalculatorServiceClient) {
	log.Println("calling sum api")
	resp, err := c.Sum(context.Background(), &calculatorpb.SumRequest{
		Num1: 5,
		Num2: 6,
	})
	if err != nil {
		log.Fatalf("call sum api err %v", err)
	}
	log.Printf("Sum api res %v\n", resp.GetResult())
}
func callSumWithDeadLine(c calculatorpb.CalculatorServiceClient, timeout time.Duration) {
	log.Println("calling sum with deadline api")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	resp, err := c.SumWithDeadLine(ctx, &calculatorpb.SumRequest{
		Num1: 5,
		Num2: 6,
	})
	if err != nil {
		if statusErr, ok := status.FromError(err); ok{
			if statusErr.Code() == codes.DeadlineExceeded {
				log.Println(" calling with api dealing Exceeded")
			}else{
				log.Fatalf("call sum with deadline api err %v", err)
			}
		}else{
			log.Fatalf("call sum with deadline api unknown err %v", err)
		}
	}
	log.Printf("Sum with deadline api res %v\n", resp.GetResult())
}
func callPND(c calculatorpb.CalculatorServiceClient) {
	log.Println("calling PND api")
	stream,err := c.PrimeNumber(context.Background(),&calculatorpb.PNRequest{
		Number: 120,
	})
	if err != io.EOF {
		log.Println("Server finishing streaming")
	}
	if err != nil {
		log.Fatalf("callPND err %v", err)
	}
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			log.Println("Server finishing stream")
			return
		}
		log.Printf("Prime number %v", resp.GetResult())
	}
}
func callAVG(c calculatorpb.CalculatorServiceClient){
	log.Println("calling AVG api")
	stream, err := c.Average(context.Background())
	if err != nil {
		log.Fatalf("call AVG err %v", err)
	}
	listReq := []calculatorpb.AVGRequest{
		calculatorpb.AVGRequest{
			Number: 6,
		},
		calculatorpb.AVGRequest{
			Number: 10,
		},
		calculatorpb.AVGRequest{
			Number: 11,
		},
		calculatorpb.AVGRequest{
			Number: 16,
		},
		calculatorpb.AVGRequest{
			Number: 26,
		},
		calculatorpb.AVGRequest{
			Number: 36,
		},
		calculatorpb.AVGRequest{
			Number: 46,
		},
		calculatorpb.AVGRequest{
			Number: 666,
		},
	}
	for _, req :=range listReq {
		err := stream.Send(&req)
		if err != nil {
			log.Fatalf("send avg err %v", err)
		}
		time.Sleep(1500*time.Millisecond)
	}
	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("rec avg err %v", err )
	}
	log.Printf("response %v", resp)
}

func callFindMax(c calculatorpb.CalculatorServiceClient){
	log.Println("calling find max..")
	stream, err := c.FindMax(context.Background())
	if err != nil {
		log.Fatalf("call AVG err %v", err)
	}

	waitc := make(chan struct{})

	go func(){
		listReq := []calculatorpb.FMRequest{
			calculatorpb.FMRequest{
				Num: 6,
			},
			calculatorpb.FMRequest{
				Num: 10,
			},
			calculatorpb.FMRequest{
				Num: 11,
			},
			calculatorpb.FMRequest{
				Num: 16,
			},
			calculatorpb.FMRequest{
				Num: 26,
			},
			calculatorpb.FMRequest{
				Num: 36,
			},
			calculatorpb.FMRequest{
				Num: 47776,
			},
			calculatorpb.FMRequest{
				Num: 666,
			},
		}
		for _, req :=range listReq {
			err := stream.Send(&req)
			if err != nil {
				log.Fatalf("send avg err %v", err)
				break
			}
			time.Sleep(1000*time.Millisecond)
		}
		stream.CloseSend()
	}()
	go func(){
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				log.Println("ending find max api")
			}
			if err != nil {
				log.Fatalf("recv find max err %v", err)
				break
			}
			log.Printf("Max : %v", resp.GetMax())
		}
		close(waitc)
	}()

	<-waitc
}
func callSquare(c calculatorpb.CalculatorServiceClient, num int32) {
	log.Println("calling square api")
	resp, err := c.Square(context.Background(), &calculatorpb.SquareRequest{
		Num: num,
	})
	if err != nil {
		log.Printf("call square api err %v\n", err)
		if errStatus, ok := status.FromError(err); ok {
			log.Printf("err msg: %v\n", errStatus.Message())
			log.Printf("err code: %v\n", errStatus.Code())
			if errStatus.Code() == codes.InvalidArgument {
				log.Printf("InvalidArgument num %v", num)
				return
			}
		}
	}
	log.Printf("Square result res %v\n", resp.GetResult())
}