package voicepeakgo

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/samber/lo"
)

func New() *Client {
	var voicepeakPath string

	switch runtime.GOOS {
	case "windows":
		voicepeakPath = "/Applications/voicepeak.app/Contents/MacOS/voicepeak"
	case "darwin":
		voicepeakPath = "/Applications/voicepeak.app/Contents/MacOS/voicepeak"
	case "linux":
		panic("Not implemented")
	default:
		panic("Unsupported OS")
	}

	return &Client{
		voicepeakPath: voicepeakPath,
	}
}

type ClinetIface interface {
	VoicepeakPath() string
	ListNarrator() ([]Narrator, error)
	ListEmotion(n *Narrator) ([]Emotion, error)
	Say(text string, outPath string, narrator *Narrator, opt ...SayOptions) error
	Help() error
}

type Client struct {
	voicepeakPath string
}

func (c *Client) VoicepeakPath() string {
	return c.voicepeakPath
}

func (c *Client) Help() error {
	cmd := exec.Command(c.voicepeakPath, "--help")
	cmd.Stderr = os.Stdout
	return cmd.Run()
}

var ErrOutOfSpeed = errors.New("speed must be between 50 and 200")
var ErrOutOfPitch = errors.New("pitch must be between -300 and 300")
var ErrInvalidEmotion = errors.New("invalid emotion")

type SayOptions struct {
	Speed        *int
	Pitch        *int
	EmotionName  *string
	EmotionValue int
}

func (impl *SayOptions) Validate(client ClinetIface, n *Narrator) error {
	if impl == nil {
		return nil
	}

	if impl.Speed != nil {
		if !(50 <= *impl.Speed && *impl.Speed <= 200) {
			return ErrOutOfSpeed
		}
	}
	if impl.Pitch != nil {
		if !(-300 <= *impl.Pitch && *impl.Pitch <= 300) {
			return ErrOutOfPitch
		}
	}

	if n != nil {
		if impl.EmotionName != nil {
			emotions, err := client.ListEmotion(n)
			if err != nil {
				return fmt.Errorf("failed to list emotion: %w", err)
			}

			if !lo.ContainsBy(emotions, func(e Emotion) bool {
				return e.name == *impl.EmotionName
			}) {
				return fmt.Errorf("invalid emotion: %w, not found %s", ErrInvalidEmotion, *impl.EmotionName)
			}
		}
	}

	return nil
}

func (c *Client) Say(text string, outPath string, narrator *Narrator, opt ...SayOptions) error {
	args := []string{
		"--say",
		text,
		"-o",
		outPath,
		"--narrator",
		narrator.String(),
	}
	if len(opt) != 0 {
		if err := opt[0].Validate(c, narrator); err != nil {
			return err
		}
		if opt[0].Speed != nil {
			args = append(args, "--speed", fmt.Sprintf("%d", *opt[0].Speed))
		}
		if opt[0].Pitch != nil {
			args = append(args, "--pitch", fmt.Sprintf("%d", *opt[0].Pitch))
		}
		if opt[0].EmotionName != nil {
			args = append(args, "--emotion", fmt.Sprintf("%s=%v", *opt[0].EmotionName, opt[0].EmotionValue))
		}
	}
	cmd := exec.Command(c.voicepeakPath, args...)
	return cmd.Run()
}

type Narrator struct {
	name string
}

func (n Narrator) String() string {
	return n.name
}

func (c *Client) ListNarrator() ([]Narrator, error) {
	cmd := exec.Command(c.voicepeakPath, "--list-narrator")
	stdout := new(bytes.Buffer)
	cmd.Stdout = stdout
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	narrators := bytes.Split(stdout.Bytes(), []byte{'\n'})

	var narratorList []Narrator
	for _, narrator := range narrators {
		if len(narrator) > 0 {
			narratorList = append(narratorList, Narrator{name: string(narrator)})
		}
	}
	return narratorList, nil
}

type Emotion struct {
	name string
}

func (e Emotion) String() string {
	return e.name
}

func (c *Client) ListEmotion(n *Narrator) ([]Emotion, error) {
	cmd := exec.Command(c.voicepeakPath, "--list-emotion", n.String())
	stdout := new(bytes.Buffer)
	cmd.Stdout = stdout

	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	emotions := bytes.Split(stdout.Bytes(), []byte{'\n'})
	var emotionList []Emotion
	for _, emotion := range emotions {
		if len(emotion) > 0 {
			emotionList = append(emotionList, Emotion{name: string(emotion)})
		}
	}
	return emotionList, nil
}
