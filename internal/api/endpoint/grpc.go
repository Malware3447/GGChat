package endpoint

import (
	"GGChat/internal/service/db"
	MyWS "GGChat/internal/websocket"
	"context"
	"fmt"
	"log"
	"time"

	pc "GGChat/internal/proto/conv-pdf"
	pd "GGChat/internal/proto/gen-doc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type gRPCMethod struct {
	Repo             *db.DbService
	WebsocketManager *MyWS.Manager
}

func NewAIChats(repo *db.DbService, wsManager *MyWS.Manager) *gRPCMethod {
	return &gRPCMethod{
		Repo:             repo,
		WebsocketManager: wsManager,
	}
}

func (a *gRPCMethod) DocGenerator(UserData string, DocPath string) (string, error) {
	conn, err := grpc.Dial("1033", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("не удалось подключиться: %v", err)
	}
	defer conn.Close()

	client := pd.NewDocumentGeneratorClient(conn)

	request := &pd.GenerateDocRequest{
		UserDataJson: UserData,
		DocUrl:       DocPath,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	log.Printf("Отправка запроса на генерацию документа...")

	response, err := client.GenerateDocument(ctx, request)
	if err != nil {
		log.Fatalf("ошибка вызова GenerateDocument: %v", err)
	}

	fmt.Printf("\n✅ Документ успешно сгенерирован!\n")
	fmt.Printf("Путь к готовому документу: %s\n", response.GetDoneDocPath())

	return response.GetDoneDocPath(), nil
}

func (a *gRPCMethod) ConPDF(DoneDocPath string) (string, error) {
	conn, err := grpc.Dial("1033", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("не удалось подключиться: %v", err)
	}
	defer conn.Close()

	client := pc.NewConvertatingToPdfClient(conn)

	request := &pc.ConvertatingToPdfRequest{
		DoneDocPath: DoneDocPath,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	log.Printf("Отправка запроса на генерацию документа...")

	response, err := client.ConvertingToPdf(ctx, request)
	if err != nil {
		log.Fatalf("ошибка вызова GenerateDocument: %v", err)
	}

	fmt.Printf("\n✅ Документ успешно сгенерирован!\n")
	fmt.Printf("Путь к готовому документу: %s\n", response.GetDonePdfPath())

	return response.GetDonePdfPath(), nil
}
