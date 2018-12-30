package packed

import (
	"github.com/blaubaer/goxr/common"
	. "github.com/onsi/gomega"
	"os"
	"testing"
)

func Test_Truncate(t *testing.T) {
	version1 := Version(1)
	offset := common.FileOffset(123)
	checksum := common.Crc64Of(versionToSeed[version1], headerPrefix, byte(version1), uint64(offset))

	brokenChecksum := make([]byte, len(checksum))
	copy(brokenChecksum, checksum)
	brokenChecksum[2] = 0

	addCase := func(name string, expectedVersion *Version, expectedTocOffset common.FileOffset, expectedSize int, inArgs ...interface{}) {
		t.Run(name, func(t *testing.T) {
			g := NewGomegaWithT(t)
			fn := tempFileWithBytesOf(inArgs...)
			defer deletePathForT(fn, t)
			{
				f, err := os.Open(fn)
				g.Expect(err).To(BeNil())
				g.Expect(f).ToNot(BeNil())
				defer closeForT(f, t)
				header, err := FindHeader(f)
				g.Expect(err).To(BeNil())
				if expectedVersion == nil {
					g.Expect(header).To(BeNil())
				} else {
					g.Expect(header).ToNot(BeNil())
					g.Expect(header.Version).To(Equal(*expectedVersion))
					g.Expect(header.TocOffset).To(Equal(expectedTocOffset))
				}
			}
			{
				err := Truncate(fn)
				g.Expect(err).To(BeNil())
			}
			{
				f, err := os.Open(fn)
				g.Expect(err).To(BeNil())
				g.Expect(f).ToNot(BeNil())
				defer closeForT(f, t)
				header, err := FindHeader(f)
				g.Expect(err).To(BeNil())
				g.Expect(header).To(BeNil())
				g.Expect(f.Close()).To(BeNil())
			}
			{
				g.Expect(fileSizeForT(fn, t)).To(Equal(int64(expectedSize)))
			}
		})
	}

	addCase("find direct at beginning", &version1, offset, 0,
		headerPrefix,
		version1,
		offset,
		checksum,
		garbage(100),
	)
	addCase("find after some garbage", &version1, offset, HeaderBufferSize*2+1,
		garbage(HeaderBufferSize*2+1),
		headerPrefix,
		version1,
		offset,
		checksum,
		garbage(100),
	)
	addCase("does not find because of unsupported version", nil, 0, headerLength,
		headerPrefix,
		Version(66),
		offset,
		checksum,
	)
	addCase("does not find because missing offset", nil, 0, headerPrefixLength+headerVersionLength,
		headerPrefix,
		version1,
	)
	addCase("does not find because missing checksum", nil, 0, headerPrefixLength+headerVersionLength+headerTocOffsetLength,
		headerPrefix,
		version1,
		offset,
	)
	addCase("does not find because of too short checksum", nil, 0, headerLength-1+10,
		headerPrefix,
		version1,
		offset,
		checksum[:headerChecksumLength-1],
		garbage(10),
	)
	addCase("does not find at because of broken checksum", nil, 0, headerLength+10,
		headerPrefix,
		version1,
		offset,
		brokenChecksum,
		garbage(10),
	)
	addCase("find after lot of distracting candidates", &version1, offset, 10+
		headerPrefixLength+headerVersionLength+headerTocOffsetLength+headerChecksumLength+10+
		headerPrefixLength+headerVersionLength+10+
		headerPrefixLength+headerVersionLength+headerTocOffsetLength+10+
		headerPrefixLength+headerVersionLength+headerTocOffsetLength+headerChecksumLength-1+10,
		garbage(10),

		headerPrefix,
		Version(66),
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
		offset,
		checksum,
	)
}
