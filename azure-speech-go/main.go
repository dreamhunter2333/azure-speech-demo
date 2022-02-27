package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/Microsoft/cognitive-services-speech-sdk-go/audio"
	"github.com/Microsoft/cognitive-services-speech-sdk-go/common"
	"github.com/Microsoft/cognitive-services-speech-sdk-go/speech"
	"github.com/urfave/cli/v2"
)

func main() {
	var endpoint string
	var subscriptionKey string
	var voiceName string
	var text string
	var outFile string

	app := &cli.App{
		Name:      "Azure Speech Cli Text To Speech",
		UsageText: "azure-speech-go -t 你好 -e https://eastasia.api.cognitive.microsoft.com/sts/v1.0/issuetoken -k xxx -v zh-CN-XiaomoNeural -o 1.wav",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "endpoint",
				Aliases:     []string{"e"},
				Usage:       "Azure speech endpoint",
				Destination: &endpoint,
			},
			&cli.StringFlag{
				Name:        "subscription-key",
				Aliases:     []string{"k"},
				Usage:       "Azure speech subscription Key",
				Destination: &subscriptionKey,
			},
			&cli.StringFlag{
				Name:        "voice",
				Aliases:     []string{"v"},
				Usage:       "Azure speech voice name, like zh-CN-XiaomoNeural, Full voice names is here https://docs.microsoft.com/zh-cn/azure/cognitive-services/speech-service/language-support#prebuilt-neural-voices",
				Destination: &voiceName,
			},
			&cli.StringFlag{
				Name:        "text",
				Aliases:     []string{"t"},
				Usage:       "input text",
				Destination: &text,
			},
			&cli.StringFlag{
				Name:        "out",
				Aliases:     []string{"o"},
				Usage:       "Out file",
				Destination: &outFile,
			},
		},
		Action: func(c *cli.Context) error {
			if len(text) == 0 {
				return errors.New("text can not be null")
			}
			config, err := speech.NewSpeechConfigFromEndpointWithSubscription(endpoint, subscriptionKey)
			if err != nil {
				fmt.Println("Got an error: ", err)
				return err
			}
			res := config.SetSpeechSynthesisVoiceName(voiceName)
			if res != nil {
				fmt.Println("Got an error: ", res)
				return res
			}
			defer config.Close()
			if len(outFile) == 0 {
				SynthesisToSpeaker(config, text)
			} else {
				SynthesisToAudioDataStream(config, text, outFile)
			}
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func synthesizeStartedHandler(event speech.SpeechSynthesisEventArgs) {
	defer event.Close()
	fmt.Println("Synthesis started.")
}

func synthesizingHandler(event speech.SpeechSynthesisEventArgs) {
	defer event.Close()
	fmt.Printf("Synthesizing, audio chunk size %d.\n", len(event.Result.AudioData))
}

func synthesizedHandler(event speech.SpeechSynthesisEventArgs) {
	defer event.Close()
	fmt.Printf("Synthesized, audio length %d.\n", len(event.Result.AudioData))
}

func cancelledHandler(event speech.SpeechSynthesisEventArgs) {
	defer event.Close()
	fmt.Println("Received a cancellation.")
}

func SynthesisToAudioDataStream(config *speech.SpeechConfig, text string, file string) error {
	speechSynthesizer, err := speech.NewSpeechSynthesizerFromConfig(config, nil)
	if err != nil {
		fmt.Println("Got an error: ", err)
		return err
	}
	defer speechSynthesizer.Close()

	speechSynthesizer.SynthesisStarted(synthesizeStartedHandler)
	speechSynthesizer.Synthesizing(synthesizingHandler)
	speechSynthesizer.SynthesisCompleted(synthesizedHandler)
	speechSynthesizer.SynthesisCanceled(cancelledHandler)

	for {
		// StartSpeakingTextAsync sends the result to channel when the synthesis starts.
		task := speechSynthesizer.StartSpeakingTextAsync(text)
		var outcome speech.SpeechSynthesisOutcome
		select {
		case outcome = <-task:
		case <-time.After(60 * time.Second):
			fmt.Println("Timed out")
			return errors.New("Timed out")
		}
		defer outcome.Close()
		if outcome.Error != nil {
			fmt.Println("Got an error: ", outcome.Error)
			return errors.New(fmt.Sprint("Got an error: ", outcome.Error))
		}

		// in most case we want to streaming receive the audio to lower the latency,
		// we can use AudioDataStream to do so.
		stream, err := speech.NewAudioDataStreamFromSpeechSynthesisResult(outcome.Result)
		defer stream.Close()
		if err != nil {
			fmt.Println("Got an error: ", err)
			return err
		}

		var all_audio []byte
		audio_chunk := make([]byte, 2048)
		for {
			n, err := stream.Read(audio_chunk)

			if err == io.EOF {
				break
			}

			all_audio = append(all_audio, audio_chunk[:n]...)
		}

		fmt.Printf("Read [%d] bytes from audio data stream.\n", len(all_audio))
	}
}

func SynthesisToSpeaker(config *speech.SpeechConfig, text string) error {
	audioConfig, err := audio.NewAudioConfigFromDefaultSpeakerOutput()
	if err != nil {
		fmt.Println("Got an error: ", err)
		return err
	}
	defer audioConfig.Close()
	speechSynthesizer, err := speech.NewSpeechSynthesizerFromConfig(config, audioConfig)
	if err != nil {
		fmt.Println("Got an error: ", err)
		return err
	}
	defer speechSynthesizer.Close()

	speechSynthesizer.SynthesisStarted(synthesizeStartedHandler)
	speechSynthesizer.Synthesizing(synthesizingHandler)
	speechSynthesizer.SynthesisCompleted(synthesizedHandler)
	speechSynthesizer.SynthesisCanceled(cancelledHandler)

	for {
		task := speechSynthesizer.SpeakTextAsync(text)
		var outcome speech.SpeechSynthesisOutcome
		select {
		case outcome = <-task:
		case <-time.After(60 * time.Second):
			fmt.Println("Timed out")
			return errors.New("Timed out")
		}
		defer outcome.Close()
		if outcome.Error != nil {
			fmt.Println("Got an error: ", outcome.Error)
			return errors.New(fmt.Sprint("Got an error: ", outcome.Error))
		}

		if outcome.Result.Reason == common.SynthesizingAudioCompleted {
			fmt.Printf("Speech synthesized to speaker for text [%s].\n", text)
		} else {
			cancellation, _ := speech.NewCancellationDetailsFromSpeechSynthesisResult(outcome.Result)
			fmt.Printf("CANCELED: Reason=%d.\n", cancellation.Reason)

			if cancellation.Reason == common.Error {
				fmt.Printf("CANCELED: ErrorCode=%d\nCANCELED: ErrorDetails=[%s]\nCANCELED: Did you update the subscription info?\n",
					cancellation.ErrorCode,
					cancellation.ErrorDetails)
			}
		}
	}
}
