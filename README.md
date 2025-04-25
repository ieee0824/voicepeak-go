# voicepeak-go

## sample

```Go
package main

import (
	"flag"
	"fmt"
	"log"

	voicepeakgo "github.com/ieee0824/voicepeak-go"
	"github.com/samber/lo"
)

func main() {
	t := flag.String("t", "", "text to say")
	flag.Parse()
	client := voicepeakgo.New()
	narrators, err := client.ListNarrator()
	if err != nil {
		fmt.Println("Error listing narrators:", err)
		return
	}

	if err := client.Say(*t, "test.wav", &narrators[1], voicepeakgo.SayOptions{
		EmotionName:  lo.ToPtr("teto-overactive"),
		EmotionValue: 100,
	}); err != nil {
		log.Fatalln(err)
	}
}
```