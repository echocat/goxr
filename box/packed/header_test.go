package packed

import (
	"bytes"
	"github.com/echocat/goxr/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Benchmark_FindHeader(b *testing.B) {
	version1 := Version(1)
	offset := common.FileOffset(123)
	checksum := common.Crc64Of(versionToSeed[version1], headerPrefix, byte(version1), uint64(offset))

	brokenChecksum := make([]byte, len(checksum))
	copy(brokenChecksum, checksum)
	brokenChecksum[2] = 0

	in := concatBytes(headerPrefix,
		garbage(10),

		headerPrefix,
		66,
		offset,
		checksum,
		garbage(10),

		headerPrefix,
		version1,
		garbage(10),

		headerPrefix,
		version1,
		offset,
		garbage(10),

		headerPrefix,
		version1,
		offset,
		checksum[:headerChecksumLength-1],
		garbage(10),

		headerPrefix,
		version1,
		versionToSeed[version1],
		garbage(10),

		headerPrefix,
		version1,
		offset,
		checksum,
	)

	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		header, err := FindHeader(bytes.NewBuffer(in))
		assert.NoError(b, err)
		assert.Equal(b, version1, header.Version)
		assert.Equal(b, offset, header.TocOffset)
	}
}

func Test_FindHeader(t *testing.T) {
	version1 := Version(1)
	offset := common.FileOffset(123)
	checksum := common.Crc64Of(versionToSeed[version1], headerPrefix, byte(version1), uint64(offset))

	brokenChecksum := make([]byte, len(checksum))
	copy(brokenChecksum, checksum)
	brokenChecksum[2] = 0

	addCase := func(name string, expectedVersion *Version, expectedOffset common.FileOffset, inArgs ...interface{}) {
		in := concatBytes(inArgs...)
		t.Run(name, func(t *testing.T) {
			header, err := FindHeader(bytes.NewBuffer(in))
			assert.NoError(t, err)
			if expectedVersion == nil {
			} else {
				assert.NotNil(t, header)
				assert.Equal(t, *expectedVersion, header.Version)
				assert.Equal(t, expectedOffset, header.TocOffset)
			}
		})
	}

	addCase("find direct at beginning", &version1, offset,
		headerPrefix,
		version1,
		offset,
		checksum,
		garbage(100),
	)
	addCase("find after some garbage", &version1, offset,
		1,
		garbage(HeaderBufferSize*2),
		headerPrefix,
		version1,
		offset,
		checksum,
		garbage(100),
	)
	addCase("does not find because of unsupported version", nil, 0,
		headerPrefix,
		66,
		offset,
		checksum,
	)
	addCase("does not find because missing offset", nil, 0,
		headerPrefix,
		version1,
	)
	addCase("does not find because missing checksum", nil, 0,
		headerPrefix,
		version1,
		offset,
	)
	addCase("does not find because of too short checksum", nil, 0,
		headerPrefix,
		version1,
		offset,
		checksum[:headerChecksumLength-1],
		garbage(10),
	)
	addCase("does not find at because of broken checksum", nil, 0,
		headerPrefix,
		version1,
		offset,
		brokenChecksum,
		garbage(10),
	)
	addCase("find after lot of distracting candidates", &version1, offset,
		garbage(10),

		headerPrefix,
		66,
		offset,
		checksum,
		garbage(10),

		headerPrefix,
		version1,
		garbage(10),

		headerPrefix,
		version1,
		offset,
		garbage(10),

		headerPrefix,
		version1,
		offset,
		checksum[:headerChecksumLength-1],
		garbage(10),

		headerPrefix,
		version1,
		versionToSeed[version1],
		garbage(10),

		headerPrefix,
		version1,
		offset,
		checksum,
	)
}
