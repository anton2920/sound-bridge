package main

import (
	"fmt"
	"unsafe"

	"github.com/anton2920/gofa/log"
	"github.com/anton2920/gofa/syscall"
)

/* From <sys/soundcard.h>. */
const (
	AFMT_S16_LE = 0x00000010 /* Little endian signed 16-bit */
	AFMT_S32_LE = 0x00001000 /* Little endian signed 32-bit */
)

const (
	Bits     = AFMT_S16_LE
	Channels = 2
	Rate     = 48000
)

var (
	SNDCTL_DSP_SETFMT   = syscall.IOWR('P', 5, uint(unsafe.Sizeof(int32(0))))
	SNDCTL_DSP_CHANNELS = syscall.IOWR('P', 6, uint(unsafe.Sizeof(int32(0))))
	SNDCTL_DSP_SPEED    = syscall.IOWR('P', 2, uint(unsafe.Sizeof(int32(0))))
)

func SetAudioParameters(fd int32, format int32, channels int32, speed int32) error {
	if err := syscall.Ioctl(fd, SNDCTL_DSP_SETFMT, unsafe.Pointer(&format)); err != nil {
		return fmt.Errorf("failed to set sample format: %v", err)
	}
	if err := syscall.Ioctl(fd, SNDCTL_DSP_CHANNELS, unsafe.Pointer(&channels)); err != nil {
		return fmt.Errorf("failed to set number of channels: %v", err)
	}
	if err := syscall.Ioctl(fd, SNDCTL_DSP_SPEED, unsafe.Pointer(&speed)); err != nil {
		return fmt.Errorf("failed to set sampling rate: %v", err)
	}
	return nil
}

func Route(indsp string, outdsp string) {
	infd, err := syscall.Open(indsp, syscall.O_RDONLY, 0)
	if err != nil {
		log.Fatalf("Failed to open input audio device: %v", err)
	}
	defer syscall.Close(infd)

	/* TODO(anton2920): query best parameters from input device. */
	if err := SetAudioParameters(infd, Bits, Channels, Rate); err != nil {
		log.Fatalf("Failed to set input device parameters: %v", err)
	}

	outfd, err := syscall.Open(outdsp, syscall.O_WRONLY, 0)
	if err != nil {
		log.Fatalf("Failed ot open output audio device: %v", err)
	}
	defer syscall.Close(outfd)

	if err := SetAudioParameters(outfd, Bits, Channels, Rate); err != nil {
		log.Fatalf("Failed to set output device parameters: %v", err)
	}

	buffer := make([]byte, 256)
	for {
		n, err := syscall.Read(infd, buffer)
		if err != nil {
			log.Fatalf("Failed to read from audio device: %v", err)
		}

		if _, err := syscall.Write(outfd, buffer[:n]); err != nil {
			log.Fatalf("Failed to write to audio device: %v", err)
		}
	}
}

func Bridge(dsp1 string, dsp2 string) {
	go Route(dsp1, dsp2)
	Route(dsp2, dsp1)
}

func main() {
	Bridge("/dev/dsp", "/dev/dsp4")
}
