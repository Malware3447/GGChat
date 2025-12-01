package endpoint

import (
	"GGChat/internal/models/chats"
	"GGChat/internal/service/db"
	MyWS "GGChat/internal/websocket"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	pc "GGChat/internal/proto/conv-pdf"
	pd "GGChat/internal/proto/gen-doc"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AIChats struct {
	Repo             *db.DbService
	WebsocketManager *MyWS.Manager
}

func NewAIChats(repo *db.DbService, wsManager *MyWS.Manager) *AIChats {
	return &AIChats{
		Repo:             repo,
		WebsocketManager: wsManager,
	}
}

func (a *AIChats) GetAllChatsAI(w http.ResponseWriter, r *http.Request) {
	UserID, ok := r.Context().Value("user_id").(int)
	if !ok {
		logrus.Warn("Ошибка получения ID пользователя из контекста")
		http.Error(w, "Error get chats", http.StatusBadRequest)
		return
	}

	response, err := a.Repo.GetAllChatsAI(context.Background(), UserID)
	if err != nil {
		logrus.Warn("ошибка получения списка чатов: ", err)
		http.Error(w, "Error get list chats", http.StatusBadRequest)
	}

	w.Header().Set("Content-Type", "application-json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (a *AIChats) DeleteChatAI(w http.ResponseWriter, r *http.Request) {
	IDStr := chi.URLParam(r, "uuid")
	ID, err := strconv.Atoi(IDStr)
	if err != nil {
		logrus.Warn("Ошибка конвертации ID: ", err)
		http.Error(w, "Error converting ID", http.StatusBadRequest)
		return
	}

	err = a.Repo.DeleteChatAI(context.Background(), ID)
	if err != nil {
		logrus.Warn("ошибка удаления чата: ", err)
		http.Error(w, "Error ", http.StatusBadRequest)
	}
	w.Header().Set("Content-Type", "application-json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(nil); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (a *AIChats) NewMessage(w http.ResponseWriter, r *http.Request) {
	UserID, ok := r.Context().Value("user_id").(int)
	if !ok {
		logrus.Warn("Ошибка получения ID пользователя из контекста")
		http.Error(w, "Error get chats", http.StatusBadRequest)
		return
	}

	body := chats.MessageAI{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		logrus.Warn("Неверное тело запроса.")
		http.Error(w, "Invflid body request", http.StatusBadRequest)
		return
	}
	body.SenderType = "user"

	go a.AIMessage(context.Background(), body.Content)

	Title, err := GenerateTitle(body.Content)
	if err != nil {
		logrus.Warn("ошибка генерации названия чата: ", err)
		http.Error(w, "ошибка генерации названия чата", http.StatusBadRequest)
	}

	if body.ChatID == 0 {
		_, err := a.Repo.CreateChatAI(context.Background(), UserID, Title)
		if err != nil {
			logrus.Warn("ошибка создания чата")
			http.Error(w, "ошибка создания чата", http.StatusBadRequest)
		}
	}

	response, err := a.Repo.CreateMessage(context.Background(), body.ChatID, body.SenderType, body.Content)
	if err != nil {
		logrus.Warn("ошибка отправки сообщения: ", err)
		http.Error(w, "ошибка отправки сообщения", http.StatusBadRequest)
	}

	w.Header().Set("Content-Type", "application-json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (a *AIChats) AIMessage(ctx context.Context, UserMessage string) {

}

func (a *AIChats) DocGenerator(UserData string, DocPath string) (string, error) {
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

func (a *AIChats) ConPDF(DoneDocPath string) (*pc.ConvertatingtoPdfResponse, error) {
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

	return response, nil
}
