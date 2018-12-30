package packed

import (
	"bufio"
	"bytes"
	"github.com/blaubaer/goxr/common"
	"hash/crc64"
	"io"
)

const (
	HeaderPrefix     = "goxr.box"
	HeaderBufferSize = 1024 * 10
)

type Version uint8

var (
	versionToSeed = map[Version][]byte{
		1: {53, 58, 197, 194, 220, 233, 145, 140, 69, 167},
	}
	headerPrefix          = []byte(HeaderPrefix)
	headerPrefixLength    = len(headerPrefix)
	headerVersionLength   = 1
	headerTocOffsetLength = 8
	headerChecksumLength  = crc64.Size
	headerLength          = headerPrefixLength + headerVersionLength + headerTocOffsetLength + headerChecksumLength
)

func WriteHeader(version Version, tocOffset common.FileOffset, to io.Writer) error {
	seed, ok := versionToSeed[version]
	if !ok {
		return ErrInvalidHeaderVersion
	}
	checksumBytes := common.Crc64Of(seed, headerPrefix, byte(version), uint64(tocOffset))

	headerBytes := common.ConcatBytes(headerPrefix, byte(version), uint64(tocOffset), checksumBytes)
	return common.Write(headerBytes, to)
}

type Header struct {
	Version   Version
	Offset    common.FileOffset
	TocOffset common.FileOffset
}

func FindHeader(r io.Reader) (header *Header, err error) {
	offset := common.FileOffset(0)
	br := bufio.NewReaderSize(r, HeaderBufferSize)
	for {
		b, err := br.ReadByte()
		if err == io.EOF {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		offset++
		if b == headerPrefix[0] {
			peeked, err := br.Peek(headerLength - 1)
			if err == io.EOF {
				return nil, nil
			}
			if err != nil {
				return nil, err
			}
			version, tocOffset, err := checkHeaderCandidate(append([]byte{b}, peeked...))
			if err != nil {
				return nil, err
			} else if version != nil {
				if _, err := br.Discard(headerLength - 1); err != nil {
					return nil, err
				}
				offset--
				return &Header{
					Version:   *version,
					Offset:    offset,
					TocOffset: tocOffset,
				}, nil
			}
		}
	}
}

func checkHeaderCandidate(candidate []byte) (*Version, common.FileOffset, error) {
	if len(candidate) != headerLength {
		return nil, 0, nil
	}
	r := bytes.NewReader(candidate)
	if actualHeaderPrefix, err := common.ReadBytes(r, headerPrefixLength); err == io.EOF {
		return nil, 0, nil
	} else if err != nil {
		return nil, 0, err
	} else if !bytes.Equal(actualHeaderPrefix, headerPrefix) {
		return nil, 0, nil
	} else if actualVersion, err := common.ReadBytes(r, headerVersionLength); err == io.EOF {
		return nil, 0, nil
	} else if err != nil {
		return nil, 0, err
	} else if seed, validVersion := versionToSeed[Version(actualVersion[0])]; !validVersion {
		return nil, 0, nil
	} else if actualTocOffset, err := common.ReadBytes(r, headerTocOffsetLength); err == io.EOF {
		return nil, 0, nil
	} else if err != nil {
		return nil, 0, err
	} else if actualChecksum, err := common.ReadBytes(r, headerChecksumLength); err == io.EOF {
		return nil, 0, nil
	} else if err != nil {
		return nil, 0, err
	} else {
		expectedChecksum := common.Crc64Of(seed, headerPrefix, actualVersion, actualTocOffset)
		if !bytes.Equal(actualChecksum, expectedChecksum) {
			return nil, 0, nil
		} else {
			version := Version(actualVersion[0])
			return &version, common.FileOffset(common.BytesToUint64(actualTocOffset)), nil
		}
	}
}
