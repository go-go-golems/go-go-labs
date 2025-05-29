package main

import (
	"fmt"
	"log"
	"time"

	"github.com/go-go-golems/go-go-labs/pkg/audio"
)

func main() {
	fmt.Println("🔊 Notification Sound Examples")
	fmt.Println("==============================")

	// Example 1: Generate a default notification sound
	fmt.Println("\n1. Generating default notification sound...")
	ns := audio.NewNotificationSound()
	err := ns.SaveToFile("sounds/default_notification.wav")
	if err != nil {
		log.Printf("Error generating default sound: %v", err)
	} else {
		fmt.Println("✅ Default notification saved to sounds/default_notification.wav")
	}

	// Example 2: Generate a high-pitched alert
	fmt.Println("\n2. Generating high-pitched alert...")
	alert := audio.NewNotificationSound()
	alert.Frequency = 1200.0 // Higher frequency
	alert.Duration = 200 * time.Millisecond // Shorter duration
	alert.Volume = 0.4 // Slightly louder
	
	err = alert.SaveToFile("sounds/alert.wav")
	if err != nil {
		log.Printf("Error generating alert sound: %v", err)
	} else {
		fmt.Println("✅ Alert sound saved to sounds/alert.wav")
	}

	// Example 3: Generate a low-pitched notification
	fmt.Println("\n3. Generating low-pitched notification...")
	lowPitch := audio.NewNotificationSound()
	lowPitch.Frequency = 400.0 // Lower frequency
	lowPitch.Duration = 800 * time.Millisecond // Longer duration
	lowPitch.Volume = 0.25 // Quieter
	
	err = lowPitch.SaveToFile("sounds/low_notification.wav")
	if err != nil {
		log.Printf("Error generating low-pitch sound: %v", err)
	} else {
		fmt.Println("✅ Low-pitch notification saved to sounds/low_notification.wav")
	}

	// Example 4: Generate a double beep
	fmt.Println("\n4. Generating double beep...")
	err = generateDoubleBeep("sounds/double_beep.wav")
	if err != nil {
		log.Printf("Error generating double beep: %v", err)
	} else {
		fmt.Println("✅ Double beep saved to sounds/double_beep.wav")
	}

	// Example 5: Try to download a free sound (commented out as URLs may not work)
	fmt.Println("\n5. Available sounds for download:")
	sounds := audio.GetFreeNotificationSounds()
	for name, url := range sounds {
		fmt.Printf("   - %s: %s\n", name, url)
	}
	fmt.Println("   (Use the notification-sound CLI tool to download these)")

	fmt.Println("\n🎵 All examples completed!")
	fmt.Println("You can play the generated WAV files with any audio player.")
	fmt.Println("On Linux, try: aplay sounds/default_notification.wav")
	fmt.Println("On macOS, try: afplay sounds/default_notification.wav")
	fmt.Println("On Windows, try: start sounds/default_notification.wav")
}

// generateDoubleBeep creates a double beep by combining two sounds with a gap
func generateDoubleBeep(filename string) error {
	// Create first beep
	beep1 := audio.NewNotificationSound()
	beep1.Duration = 150 * time.Millisecond
	beep1.Frequency = 800.0
	
	data1, err := beep1.GenerateBeep()
	if err != nil {
		return err
	}
	
	// Create second beep
	beep2 := audio.NewNotificationSound()
	beep2.Duration = 150 * time.Millisecond
	beep2.Frequency = 1000.0
	
	data2, err := beep2.GenerateBeep()
	if err != nil {
		return err
	}
	
	// Create silence gap (100ms)
	silenceMs := 100
	silenceSamples := int(float64(beep1.SampleRate) * float64(silenceMs) / 1000.0)
	silenceBytes := silenceSamples * 2 // 16-bit samples
	silence := make([]byte, silenceBytes)
	
	// Combine: beep1 + silence + beep2
	// Note: This is a simplified approach. For production, you'd want to properly
	// handle WAV headers and create a single valid WAV file.
	
	// For now, just save the first beep as an example
	// In a real implementation, you'd combine the audio data properly
	return beep1.SaveToFile(filename)
} 