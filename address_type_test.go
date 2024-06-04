package diameter_test

import (
	"net"

	"github.com/blorticus-go/diameter"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AddressType", func() {
	Describe("creating AddressType using NewAddressTypeErrorable()", func() {
		When("passing IP4 with 4 byte slice", func() {
			var addressType diameter.AddressType
			var err error

			BeforeEach(func() {
				addressType, err = diameter.NewAddressTypeErrorable(diameter.IP4, []byte{10, 254, 10, 1})
			})

			It("does not raise an error", func() {
				Expect(err).To(BeNil())
			})

			It("encodes properly", func() {
				Expect([]byte(addressType)).To(Equal([]byte{0x0, 0x1, 10, 254, 10, 1}))
			})

			It("reports the Type() as IP4", func() {
				Expect(addressType.Type()).To(Equal(diameter.IP4))
			})

			It("reports the Address() as the byte array", func() {
				Expect(addressType.Address()).To(Equal([]byte{10, 254, 10, 1}))
			})

			It("returns true for IsAnIP()", func() {
				Expect(addressType.IsAnIP()).To(BeTrue())
			})

			It("returns false for IsNotAnIP()", func() {
				Expect(addressType.IsNotAnIP()).To(BeFalse())
			})

			It("returns the corresponding net.IP object from ToIP()", func() {
				Expect(addressType.ToIP().Equal(net.ParseIP("10.254.10.1"))).To(BeTrue())
			})
		})

		When("passing IP6 with a 16 byte slice", func() {
			var addressType diameter.AddressType
			var err error

			BeforeEach(func() {
				addressType, err = diameter.NewAddressTypeErrorable(diameter.IP6, []byte{0xfd, 0x00, 0xab, 0xcd, 0x00, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0, 0x00, 0x01})
			})

			It("does not raise an error", func() {
				Expect(err).To(BeNil())
			})

			It("encodes properly", func() {
				Expect([]byte(addressType)).To(Equal([]byte{0x0, 0x2, 0xfd, 0x00, 0xab, 0xcd, 0x00, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0, 0x00, 0x01}))
			})

			It("reports the Type() as IP6", func() {
				Expect(addressType.Type()).To(Equal(diameter.IP6))
			})

			It("reports the Address() as the byte array", func() {
				Expect(addressType.Address()).To(Equal([]byte{0xfd, 0x00, 0xab, 0xcd, 0x00, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0, 0x00, 0x01}))
			})

			It("returns true for IsAnIP()", func() {
				Expect(addressType.IsAnIP()).To(BeTrue())
			})

			It("returns false for IsNotAnIP()", func() {
				Expect(addressType.IsNotAnIP()).To(BeFalse())
			})

			It("returns the corresponding net.IP object from ToIP()", func() {
				Expect(addressType.ToIP().Equal(net.ParseIP("fd00:abcd:0:1::1"))).To(BeTrue())
			})
		})

		When("passing IP4 with a 16 byte slice", func() {
			var err error

			BeforeEach(func() {
				_, err = diameter.NewAddressTypeErrorable(diameter.IP4, []byte{0xfd, 0x00, 0xab, 0xcd, 0x00, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0, 0x00, 0x01})
			})

			It("raises an error", func() {
				Expect(err).ToNot(BeNil())
			})
		})

		When("passing IP6 with a 4 byte slice", func() {
			var err error

			BeforeEach(func() {
				_, err = diameter.NewAddressTypeErrorable(diameter.IP6, []byte{0xfd, 0x00, 0xab, 0xcd})
			})

			It("raises an error", func() {
				Expect(err).ToNot(BeNil())
			})
		})

		When("passing MAC48Bit with 6 byte slice", func() {
			var addressType diameter.AddressType
			var err error

			BeforeEach(func() {
				addressType, err = diameter.NewAddressTypeErrorable(diameter.MAC48Bit, []byte{0x00, 0x10, 0xff, 0x23, 0xee, 0x45})
			})

			It("does not raise an error", func() {
				Expect(err).To(BeNil())
			})

			It("encodes properly", func() {
				Expect([]byte(addressType)).To(Equal([]byte{0x40, 0x05, 0x00, 0x10, 0xff, 0x23, 0xee, 0x45}))
			})

			It("reports the Type() as MAC48Bit", func() {
				Expect(addressType.Type()).To(Equal(diameter.MAC48Bit))
			})

			It("reports the Address() as the byte array", func() {
				Expect(addressType.Address()).To(Equal([]byte{0x00, 0x10, 0xff, 0x23, 0xee, 0x45}))
			})

			It("returns false for IsAnIP()", func() {
				Expect(addressType.IsAnIP()).To(BeFalse())
			})

			It("returns true for IsNotAnIP()", func() {
				Expect(addressType.IsNotAnIP()).To(BeTrue())
			})

			It("returns the nil from ToIP()", func() {
				Expect(addressType.ToIP()).To(BeNil())
			})
		})

	})
})
