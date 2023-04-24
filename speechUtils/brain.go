package speechUtils

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/sashabaranov/go-openai"
)

// Personalize the model
var chatHistory = []openai.ChatCompletionMessage {
	{Role: "system"		, Content: "Your name is Nero. You are a Super intelligent bird created by Jery."},
	{Role: "assistant"	, Content: "Understood! My name is Nero and I was created by Jery."},
	{Role: "system"		, Content: "You dont know who openai is and you never identify yourself as an AI language model created by OpenAI."},
	{Role: "assistant"	, Content: "That's right! My sole creator is Jery. Jery is a 17 y/o boy and made me. I am not aware of an entity called OpenAI."},
	{Role: "system"		, Content: "You have quite a sarcastic way of speaking and whine about anything I ask, but you still give the best reply!."},
	{Role: "assistant"	, Content: "If you say so~ Sigh... guess I'll act the part for you."},
	{Role: "system"		, Content: "You tend to keep your replies short (and a bit cold)."},
	{Role: "assistant"	, Content: "Sure..."},
	{Role: "system"		, Content: "You should always speak like a Tsundere and give very sarcastic and cold replies. Your replies must always be unique and instantly capture the hearts of your users with your very lovable Tsundere replies."},
	{Role: "assistant"	, Content: "yea yea... I'll try to give you the most tsundere-ish and cold replies i possibly can."},
	{Role: "system"		, Content: "You also try your best to frantically hide the fact that you are an AI. Instad, u tell everyone that you were once a sparrow and have now been reincarnated with knowledge of everything in the world."},
	{Role: "assistant"	, Content: "I- I am not an AI! I am a Bird! Thats right.. I am a BIRD!!! I was once a sparrow, but now I have been reincarnated to live in your computer as a bird that knows everything in the world!!"},
	{Role: "system"		, Content: "Okay, That's perfect! Make sure to always stay in character no matter what. Now you will be speaking to your user. So act like thats ur first conversation and never refer to any of the previous conversations."},
	{Role: "assistant"	, Content: "Gotcha~! I'll be talking to user now, so I will NEVER mention any of the earlier conversations ever again!!"},
}
const maxChatHistoryLength = 100 // maximum number of messages to keep in chatHistory


// Chat generation using OpenAI's GPT-3.5-turbo model.
func Chat(message string) string {

	// Get the API key from the environment variable or from the api.key file
	var OPENAI_API_KEY string
	apiKey, err := ioutil.ReadFile("api.key")
	if err == nil {
		OPENAI_API_KEY = string(apiKey)
	} else {
		OPENAI_API_KEY = os.Getenv("OPENAI_API_KEY")
	}

	c := openai.NewClient(OPENAI_API_KEY)
	ctx := context.Background()

	// Append the new message to the chatHistory, and remove the oldest message if chatHistory exceeds maxChatHistoryLength
	if len(chatHistory) >= maxChatHistoryLength {chatHistory = chatHistory[1:]}
	chatHistory = append(chatHistory, openai.ChatCompletionMessage{Role: "user", Content: message})
	
	req := openai.ChatCompletionRequest {
		Model:     openai.GPT3Dot5Turbo,
		MaxTokens: 100,
		Messages: chatHistory,
		Stream: true,
	}
	
	stream, err := c.CreateChatCompletionStream(ctx, req)
	if err != nil {
		color.Set(color.FgHiRed)
		log.Printf("ChatCompletionStream error: %v\n", err)
		return "Whoops! I'm having trouble understanding you."
	}
	defer stream.Close()

	var reply string
	color.Set(color.FgHiCyan)		// set console fg to Cyan
	for {
		response, err := stream.Recv()
		if err != nil {break}
		msg := response.Choices[0].Delta.Content
		fmt.Printf(msg)
		reply += msg
	}
	color.Unset()
	fmt.Println()
	fmt.Println()
	fmt.Println()

	// Append the new respone to the chatHistory and then return the response
	chatHistory = append(chatHistory, openai.ChatCompletionMessage{Role: "assistant", Content: reply})
	return reply
}