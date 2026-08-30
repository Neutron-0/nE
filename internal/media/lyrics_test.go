package media_test

import (
	"testing"

	"ne/internal/media"
)

func TestParseLRC(t *testing.T) {
	sampleLRC := `[ti:Moonlight]
[ar:Kali Uchis]
[al:Red Moon in Venus]
[00:08.50]I, I just wanna get high with my lover
[00:15.80]Veo una muñeca cuando me miro en el espejo
[00:23.20]Kiss, kiss, looking so good, it's to die for
[00:30.00]Riding round town, they gon' feel this one
`

	lines := media.ParseLRC(sampleLRC)
	if len(lines) != 4 {
		t.Fatalf("expected 4 lyric lines, got %d", len(lines))
	}

	if lines[0].Time != 8.5 || lines[0].Text != "I, I just wanna get high with my lover" {
		t.Errorf("line 0 mismatch: time=%.2f, text=%q", lines[0].Time, lines[0].Text)
	}

	if lines[1].Time != 15.8 || lines[1].Text != "Veo una muñeca cuando me miro en el espejo" {
		t.Errorf("line 1 mismatch: time=%.2f, text=%q", lines[1].Time, lines[1].Text)
	}

	if lines[3].Time != 30.0 {
		t.Errorf("line 3 time mismatch: got %.2f, expected 30.00", lines[3].Time)
	}
}
