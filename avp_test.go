package diameter_test

import (
	"net"
	"time"

	"github.com/blorticus-go/diameter"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AVP", func() {
	Describe("creating new untyped AVPs", func() {
		When("creating Origin-Host", func() {
			avp := diameter.NewAVP(264, 0, true, []byte("client.example.com"))

			It("properly sets AVP exported fields", func() {
				Expect(avp).To(Equal(&diameter.AVP{
					Code:               264,
					VendorID:           0,
					VendorSpecific:     false,
					Mandatory:          true,
					Length:             26,
					PaddedLength:       28,
					Data:               []byte("client.example.com"),
					ExtendedAttributes: nil,
				}))
			})

			It("properly Encodes", func() {
				Expect(avp.Encode()).To(Equal([]byte{
					0x00, 0x00, 0x01, 0x08,
					0x40, 0x00, 0x00, 0x1a,
					0x63, 0x6c, 0x69, 0x65,
					0x6e, 0x74, 0x2e, 0x65,
					0x78, 0x61, 0x6d, 0x70,
					0x6c, 0x65, 0x2e, 0x63,
					0x6f, 0x6d, 0x00, 0x00,
				}))
			})
		})
	})

	Describe("creating typed AVPs", func() {
		Describe("creating Origin-Host (type DiamIdent)", func() {
			When("creating with value 'client.example.com'", func() {
				var avp *diameter.AVP
				var err error
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(264, 0, true, diameter.DiamIdent, "client.example.com")
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           264,
						VendorID:       0,
						VendorSpecific: false,
						Mandatory:      true,
						Length:         26,
						PaddedLength:   28,
						Data:           []byte("client.example.com"),
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.DiamIdent,
							TypedValue: "client.example.com",
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0x00, 0x01, 0x08,
						0x40, 0x00, 0x00, 0x1a,
						0x63, 0x6c, 0x69, 0x65,
						0x6e, 0x74, 0x2e, 0x65,
						0x78, 0x61, 0x6d, 0x70,
						0x6c, 0x65, 0x2e, 0x63,
						0x6f, 0x6d, 0x00, 0x00,
					}))
				})
			})

			When("creating with value '' (the empty string)", func() {
				var avp *diameter.AVP
				var err error
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(264, 0, true, diameter.DiamIdent, "")
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           264,
						VendorID:       0,
						VendorSpecific: false,
						Mandatory:      true,
						Length:         8,
						PaddedLength:   8,
						Data:           []byte(""),
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.DiamIdent,
							TypedValue: "",
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0x00, 0x01, 0x08,
						0x40, 0x00, 0x00, 0x08,
					}))
				})

			})

			When("using a raw bytes slice for data", func() {
				var err error
				BeforeEach(func() {
					_, err = diameter.NewTypedAVPErrorable(264, 0, true, diameter.DiamIdent, []byte("client.example.com"))
				})

				It("returns an error", func() {
					Expect(err).ToNot(BeNil())
				})
			})

			When("a nil value for data", func() {
				var err error
				BeforeEach(func() {
					_, err = diameter.NewTypedAVPErrorable(264, 0, true, diameter.DiamIdent, nil)
				})

				It("returns an error", func() {
					Expect(err).ToNot(BeNil())
				})
			})

		})

		Describe("creating Redirect-Host (type DiamURI)", func() {
			When("using the string 'aaa://host.example.com'", func() {
				var avp *diameter.AVP
				var err error
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(292, 0, true, diameter.DiamURI, "aaa://host.example.com")
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           292,
						VendorID:       0,
						VendorSpecific: false,
						Mandatory:      true,
						Length:         30,
						PaddedLength:   32,
						Data:           []byte("aaa://host.example.com"),
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.DiamURI,
							TypedValue: "aaa://host.example.com",
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0x00, 0x01, 0x24,
						0x40, 0x00, 0x00, 0x1e,
						0x61, 0x61, 0x61, 0x3a,
						0x2f, 0x2f, 0x68, 0x6f,
						0x73, 0x74, 0x2e, 0x65,
						0x78, 0x61, 0x6d, 0x70,
						0x6c, 0x65, 0x2e, 0x63,
						0x6f, 0x6d, 0x00, 0x00,
					}))
				})
			})

			When("using the string '' (the empty string) -- technically, this is illegal, but AVP check for string doesn't validate", func() {
				var avp *diameter.AVP
				var err error
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(292, 0, true, diameter.DiamURI, "")
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           292,
						VendorID:       0,
						VendorSpecific: false,
						Mandatory:      true,
						Length:         8,
						PaddedLength:   8,
						Data:           []byte(""),
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.DiamURI,
							TypedValue: "",
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0x00, 0x01, 0x24,
						0x40, 0x00, 0x00, 0x08,
					}))
				})

			})

			When("using a raw byte slice", func() {
				var err error
				BeforeEach(func() {
					_, err = diameter.NewTypedAVPErrorable(292, 0, true, diameter.DiamURI, []byte(""))
				})

				It("returns an error", func() {
					Expect(err).ToNot(BeNil())
				})
			})

			When("using a nil value", func() {
				var err error
				BeforeEach(func() {
					_, err = diameter.NewTypedAVPErrorable(292, 0, true, diameter.DiamURI, nil)
				})

				It("returns an error", func() {
					Expect(err).ToNot(BeNil())
				})
			})

		})

		Describe("creating Result-Code (type Unsigned32)", func() {
			When("using a value of uint32(2001)", func() {
				var err error
				var avp *diameter.AVP
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(268, 0, true, diameter.Unsigned32, uint32(2001))
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           268,
						VendorID:       0,
						VendorSpecific: false,
						Mandatory:      true,
						Length:         12,
						PaddedLength:   12,
						Data:           []byte{0x00, 0x00, 0x07, 0xd1},
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.Unsigned32,
							TypedValue: uint32(2001),
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0x00, 0x01, 0x0c,
						0x40, 0x00, 0x00, 0x0c,
						0x00, 0x00, 0x07, 0xd1,
					}))
				})
			})

			When("using a value of uint32(0)", func() {
				var err error
				var avp *diameter.AVP
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(268, 0, true, diameter.Unsigned32, uint32(0))
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           268,
						VendorID:       0,
						VendorSpecific: false,
						Mandatory:      true,
						Length:         12,
						PaddedLength:   12,
						Data:           []byte{0x00, 0x00, 0x00, 0x00},
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.Unsigned32,
							TypedValue: uint32(0),
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0x00, 0x01, 0x0c,
						0x40, 0x00, 0x00, 0x0c,
						0x00, 0x00, 0x00, 0x00,
					}))
				})
			})

			When("using a value of uint32(0xffffffff)", func() {
				var err error
				var avp *diameter.AVP
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(268, 0, true, diameter.Unsigned32, uint32(0xffffffff))
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           268,
						VendorID:       0,
						VendorSpecific: false,
						Mandatory:      true,
						Length:         12,
						PaddedLength:   12,
						Data:           []byte{0xff, 0xff, 0xff, 0xff},
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.Unsigned32,
							TypedValue: uint32(0xffffffff),
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0x00, 0x01, 0x0c,
						0x40, 0x00, 0x00, 0x0c,
						0xff, 0xff, 0xff, 0xff,
					}))
				})

			})

			When("using a value of uint64(2001)", func() {
				var err error
				BeforeEach(func() {
					_, err = diameter.NewTypedAVPErrorable(268, 0, true, diameter.Unsigned32, uint64(2001))
				})

				It("returns an error", func() {
					Expect(err).ToNot(BeNil())
				})
			})

			When("using a value of uint(2001)", func() {
				var err error
				BeforeEach(func() {
					_, err = diameter.NewTypedAVPErrorable(268, 0, true, diameter.Unsigned32, uint(2001))
				})

				It("returns an error", func() {
					Expect(err).ToNot(BeNil())
				})
			})

			When("using a value of int(2001)", func() {
				var err error
				var avp *diameter.AVP
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(268, 0, true, diameter.Unsigned32, int(2001))
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           268,
						VendorID:       0,
						VendorSpecific: false,
						Mandatory:      true,
						Length:         12,
						PaddedLength:   12,
						Data:           []byte{0x00, 0x00, 0x07, 0xd1},
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.Unsigned32,
							TypedValue: uint32(2001),
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0x00, 0x01, 0x0c,
						0x40, 0x00, 0x00, 0x0c,
						0x00, 0x00, 0x07, 0xd1,
					}))
				})
			})

			When("using a value of int32(2001)", func() {
				var err error
				BeforeEach(func() {
					_, err = diameter.NewTypedAVPErrorable(268, 0, true, diameter.Unsigned32, int32(2001))
				})

				It("returns an error", func() {
					Expect(err).ToNot(BeNil())
				})
			})

			When("using a value of int64(2001)", func() {
				var err error
				BeforeEach(func() {
					_, err = diameter.NewTypedAVPErrorable(268, 0, true, diameter.Unsigned32, int64(2001))
				})

				It("returns an error", func() {
					Expect(err).ToNot(BeNil())
				})
			})

			When("using a value of int(-2001)", func() {
				var err error
				var avp *diameter.AVP
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(268, 0, true, diameter.Unsigned32, int(-2001))
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           268,
						VendorID:       0,
						VendorSpecific: false,
						Mandatory:      true,
						Length:         12,
						PaddedLength:   12,
						Data:           []byte{0xff, 0xff, 0xf8, 0x2f},
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.Unsigned32,
							TypedValue: uint32(0xfffff82f),
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0x00, 0x01, 0x0c,
						0x40, 0x00, 0x00, 0x0c,
						0xff, 0xff, 0xf8, 0x2f,
					}))
				})
			})

			When("using a value of '2001' (a string)", func() {
				var err error
				BeforeEach(func() {
					_, err = diameter.NewTypedAVPErrorable(268, 0, true, diameter.Unsigned32, "2001")
				})

				It("returns an error", func() {
					Expect(err).ToNot(BeNil())
				})
			})

			When("using a value using a byte slice", func() {
				var err error
				BeforeEach(func() {
					_, err = diameter.NewTypedAVPErrorable(268, 0, true, diameter.Unsigned32, []byte{0x00, 0x00, 0x07, 0xd1})
				})

				It("returns an error", func() {
					Expect(err).ToNot(BeNil())
				})

			})
		})

		Describe("creating Accounting-Sub-Session-Id (type Unsigned64)", func() {
			When("using a value of uint64(0xff00ff00ff00ff00)", func() {
				var err error
				var avp *diameter.AVP
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(287, 0, false, diameter.Unsigned64, uint64(0xff00ff00ff00ff00))
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           287,
						VendorID:       0,
						VendorSpecific: false,
						Mandatory:      false,
						Length:         16,
						PaddedLength:   16,
						Data:           []byte{0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00},
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.Unsigned64,
							TypedValue: uint64(0xff00ff00ff00ff00),
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0x00, 0x01, 0x1f,
						0x00, 0x00, 0x00, 0x10,
						0xff, 0x00, 0xff, 0x00,
						0xff, 0x00, 0xff, 0x00,
					}))
				})
			})

			When("using a value of int(65536)", func() {
				var err error
				var avp *diameter.AVP
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(287, 0, false, diameter.Unsigned64, int(65536))
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           287,
						VendorID:       0,
						VendorSpecific: false,
						Mandatory:      false,
						Length:         16,
						PaddedLength:   16,
						Data:           []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00},
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.Unsigned64,
							TypedValue: uint64(65536),
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0x00, 0x01, 0x1f,
						0x00, 0x00, 0x00, 0x10,
						0x00, 0x00, 0x00, 0x00,
						0x00, 0x01, 0x00, 0x00,
					}))
				})
			})

			When("using a value of uint(65536)", func() {
				var err error
				var avp *diameter.AVP
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(287, 0, false, diameter.Unsigned64, uint(65536))
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           287,
						VendorID:       0,
						VendorSpecific: false,
						Mandatory:      false,
						Length:         16,
						PaddedLength:   16,
						Data:           []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00},
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.Unsigned64,
							TypedValue: uint64(65536),
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0x00, 0x01, 0x1f,
						0x00, 0x00, 0x00, 0x10,
						0x00, 0x00, 0x00, 0x00,
						0x00, 0x01, 0x00, 0x00,
					}))
				})
			})

			When("using a value of uint32(65536)", func() {
				var err error
				var avp *diameter.AVP
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(287, 0, false, diameter.Unsigned64, uint32(65536))
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           287,
						VendorID:       0,
						VendorSpecific: false,
						Mandatory:      false,
						Length:         16,
						PaddedLength:   16,
						Data:           []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00},
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.Unsigned64,
							TypedValue: uint64(65536),
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0x00, 0x01, 0x1f,
						0x00, 0x00, 0x00, 0x10,
						0x00, 0x00, 0x00, 0x00,
						0x00, 0x01, 0x00, 0x00,
					}))
				})
			})

			When("using a value of int32(65536)", func() {
				var err error
				BeforeEach(func() {
					_, err = diameter.NewTypedAVPErrorable(287, 0, false, diameter.Unsigned64, int32(65536))
				})

				It("returns an error", func() {
					Expect(err).ToNot(BeNil())
				})
			})

			When("using a value of int64(0x7fffffffffffffff)", func() {
				var err error
				BeforeEach(func() {
					_, err = diameter.NewTypedAVPErrorable(287, 0, false, diameter.Unsigned64, int64(0x7fffffffffffffff))
				})

				It("returns an error", func() {
					Expect(err).ToNot(BeNil())
				})
			})

			When("using a value of int64(-10) (which causes type converstion to a large positive integer)", func() {
				var err error
				BeforeEach(func() {
					_, err = diameter.NewTypedAVPErrorable(287, 0, false, diameter.Unsigned64, int64(-10))
				})

				It("returns an error", func() {
					Expect(err).ToNot(BeNil())
				})
			})

			When("using a value of nil", func() {
				var err error
				BeforeEach(func() {
					_, err = diameter.NewTypedAVPErrorable(287, 0, false, diameter.Unsigned64, nil)
				})

				It("returns an error", func() {
					Expect(err).ToNot(BeNil())
				})
			})

			When("using a value of '10' (string)", func() {
				var err error
				BeforeEach(func() {
					_, err = diameter.NewTypedAVPErrorable(287, 0, false, diameter.Unsigned64, "10")
				})

				It("returns an error", func() {
					Expect(err).ToNot(BeNil())
				})
			})
		})

		Describe("creating Exponent (type Integer32)", func() {
			When("using a value of int32(0)", func() {
				var err error
				var avp *diameter.AVP
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(429, 0, true, diameter.Integer32, int32(0))
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           429,
						VendorID:       0,
						VendorSpecific: false,
						Mandatory:      true,
						Length:         12,
						PaddedLength:   12,
						Data:           []byte{0x00, 0x00, 0x00, 0x00},
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.Integer32,
							TypedValue: int32(0),
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0x00, 0x01, 0xad,
						0x40, 0x00, 0x00, 0x0c,
						0x00, 0x00, 0x00, 0x00,
					}))
				})
			})

			When("using a value of int32(-43201652)", func() {
				var err error
				var avp *diameter.AVP
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(429, 0, true, diameter.Integer32, int32(-43201652))
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           429,
						VendorID:       0,
						VendorSpecific: false,
						Mandatory:      true,
						Length:         12,
						PaddedLength:   12,
						Data:           []byte{0xfd, 0x6c, 0xcb, 0x8c},
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.Integer32,
							TypedValue: int32(-43201652),
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0x00, 0x01, 0xad,
						0x40, 0x00, 0x00, 0x0c,
						0xfd, 0x6c, 0xcb, 0x8c,
					}))
				})
			})

			When("using a value of int32(43201652)", func() {
				var err error
				var avp *diameter.AVP
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(429, 0, true, diameter.Integer32, int32(43201652))
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           429,
						VendorID:       0,
						VendorSpecific: false,
						Mandatory:      true,
						Length:         12,
						PaddedLength:   12,
						Data:           []byte{0x02, 0x93, 0x34, 0x74},
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.Integer32,
							TypedValue: int32(43201652),
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0x00, 0x01, 0xad,
						0x40, 0x00, 0x00, 0x0c,
						0x02, 0x93, 0x34, 0x74,
					}))
				})
			})

			When("using a value of int(65536)", func() {
				var err error
				var avp *diameter.AVP
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(429, 0, true, diameter.Integer32, int(65536))
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           429,
						VendorID:       0,
						VendorSpecific: false,
						Mandatory:      true,
						Length:         12,
						PaddedLength:   12,
						Data:           []byte{0x00, 0x01, 0x00, 0x00},
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.Integer32,
							TypedValue: int32(65536),
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0x00, 0x01, 0xad,
						0x40, 0x00, 0x00, 0x0c,
						0x00, 0x01, 0x00, 0x00,
					}))
				})
			})

			When("using a value of int64(0x7fffffff70f0f0f0)", func() {
				var err error
				BeforeEach(func() {
					_, err = diameter.NewTypedAVPErrorable(429, 0, true, diameter.Integer32, int64(0x7fffffff70f0f0f0))
				})

				It("returns an error", func() {
					Expect(err).ToNot(BeNil())
				})
			})

			When("using a value of uint64(0x7fffffff70f0f0f0) -- will truncate to 32 bits", func() {
				var err error
				BeforeEach(func() {
					_, err = diameter.NewTypedAVPErrorable(429, 0, true, diameter.Integer32, uint64(0x7fffffff70f0f0f0))
				})

				It("returns an error", func() {
					Expect(err).ToNot(BeNil())
				})
			})

			When("using a value of uint32(0xffffffff)", func() {
				var err error
				BeforeEach(func() {
					_, err = diameter.NewTypedAVPErrorable(429, 0, true, diameter.Integer32, uint32(0xffffffff))
				})

				It("returns an error", func() {
					Expect(err).ToNot(BeNil())
				})
			})

			When("using a value of uint(0xffffffff)", func() {
				var err error
				BeforeEach(func() {
					_, err = diameter.NewTypedAVPErrorable(429, 0, true, diameter.Integer32, uint(0xffffffff))
				})

				It("returns an error", func() {
					Expect(err).ToNot(BeNil())
				})
			})

			When("using a value of nil", func() {
				var err error
				BeforeEach(func() {
					_, err = diameter.NewTypedAVPErrorable(429, 0, true, diameter.Integer32, nil)
				})

				It("returns an error", func() {
					Expect(err).ToNot(BeNil())
				})
			})

			When("using a value of '10' (a string)", func() {
				var err error
				BeforeEach(func() {
					_, err = diameter.NewTypedAVPErrorable(429, 0, true, diameter.Integer32, "10")
				})

				It("returns an error", func() {
					Expect(err).ToNot(BeNil())
				})
			})

		})

		Describe("creating Value-Digits (type Integer64)", func() {
			When("using a value of int32(0)", func() {
				var err error
				var avp *diameter.AVP
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(447, 0, true, diameter.Integer64, int64(0))
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           447,
						VendorID:       0,
						VendorSpecific: false,
						Mandatory:      true,
						Length:         16,
						PaddedLength:   16,
						Data:           []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.Integer64,
							TypedValue: int64(0),
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0x00, 0x01, 0xbf,
						0x40, 0x00, 0x00, 0x10,
						0x00, 0x00, 0x00, 0x00,
						0x00, 0x00, 0x00, 0x00,
					}))
				})
			})

			When("using a value of int64(-987654321000)", func() {
				var err error
				var avp *diameter.AVP
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(447, 0, true, diameter.Integer64, int64(-987654321000))
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           447,
						VendorID:       0,
						VendorSpecific: false,
						Mandatory:      true,
						Length:         16,
						PaddedLength:   16,
						Data:           []byte{0xff, 0xff, 0xff, 0x1a, 0x0b, 0x37, 0x0c, 0x98},
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.Integer64,
							TypedValue: int64(-987654321000),
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0x00, 0x01, 0xbf,
						0x40, 0x00, 0x00, 0x10,
						0xff, 0xff, 0xff, 0x1a,
						0x0b, 0x37, 0x0c, 0x98,
					}))
				})
			})

			When("using a value of int32(43201652)", func() {
				var err error
				var avp *diameter.AVP
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(447, 0, true, diameter.Integer64, int32(43201652))
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           447,
						VendorID:       0,
						VendorSpecific: false,
						Mandatory:      true,
						Length:         16,
						PaddedLength:   16,
						Data:           []byte{0, 0, 0, 0, 0x02, 0x93, 0x34, 0x74},
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.Integer64,
							TypedValue: int64(43201652),
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0x00, 0x01, 0xbf,
						0x40, 0x00, 0x00, 0x10,
						0x00, 0x00, 0x00, 0x00,
						0x02, 0x93, 0x34, 0x74,
					}))
				})
			})

			When("using a value of int(65536)", func() {
				var err error
				var avp *diameter.AVP
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(447, 0, true, diameter.Integer64, int(65536))
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           447,
						VendorID:       0,
						VendorSpecific: false,
						Mandatory:      true,
						Length:         16,
						PaddedLength:   16,
						Data:           []byte{0, 0, 0, 0, 0x00, 0x01, 0x00, 0x00},
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.Integer64,
							TypedValue: int64(65536),
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0x00, 0x01, 0xbf,
						0x40, 0x00, 0x00, 0x10,
						0x00, 0x00, 0x00, 0x00,
						0x00, 0x01, 0x00, 0x00,
					}))
				})

			})

			When("using a value of uint64(0xffffffffffffffff)", func() {
				var err error
				BeforeEach(func() {
					_, err = diameter.NewTypedAVPErrorable(447, 0, true, diameter.Integer64, uint64(0xffffffffffffffff))
				})

				It("returns an error", func() {
					Expect(err).ToNot(BeNil())
				})
			})

			When("using a value of uint32(0xffffffff)", func() {
				var err error
				BeforeEach(func() {
					_, err = diameter.NewTypedAVPErrorable(447, 0, true, diameter.Integer64, uint32(0xffffffff))
				})

				It("returns an error", func() {
					Expect(err).ToNot(BeNil())
				})
			})

			When("using a value of uint(0xffffffff)", func() {
				var err error
				BeforeEach(func() {
					_, err = diameter.NewTypedAVPErrorable(447, 0, true, diameter.Integer64, uint(0xffffffff))
				})

				It("returns an error", func() {
					Expect(err).ToNot(BeNil())
				})
			})

			When("using a value of nil", func() {
				var err error
				BeforeEach(func() {
					_, err = diameter.NewTypedAVPErrorable(447, 0, true, diameter.Integer64, nil)
				})

				It("returns an error", func() {
					Expect(err).ToNot(BeNil())
				})
			})

			When("using a value of '10' (a string)", func() {
				var err error
				BeforeEach(func() {
					_, err = diameter.NewTypedAVPErrorable(447, 0, true, diameter.Integer64, "10")
				})

				It("returns an error", func() {
					Expect(err).ToNot(BeNil())
				})
			})

		})

		Describe("creating custom AVP 16777216:100 (type Float32)", func() {
			When("using a value of float32(0)", func() {
				var err error
				var avp *diameter.AVP
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(100, 16777216, false, diameter.Float32, float32(0))
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           100,
						VendorID:       16777216,
						VendorSpecific: true,
						Mandatory:      false,
						Length:         16,
						PaddedLength:   16,
						Data:           []byte{0, 0, 0, 0},
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.Float32,
							TypedValue: float32(0),
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0x00, 0x00, 0x64,
						0x80, 0x00, 0x00, 0x10,
						0x01, 0x00, 0x00, 0x00,
						0x00, 0x00, 0x00, 0x00,
					}))
				})
			})

			When("using a value of float32(1234.5678)", func() {
				var err error
				var avp *diameter.AVP
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(100, 16777216, false, diameter.Float32, float32(1234.5678))
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           100,
						VendorID:       16777216,
						VendorSpecific: true,
						Mandatory:      false,
						Length:         16,
						PaddedLength:   16,
						Data:           []byte{0x44, 0x9a, 0x52, 0x2b},
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.Float32,
							TypedValue: float32(1234.5678),
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0x00, 0x00, 0x64,
						0x80, 0x00, 0x00, 0x10,
						0x01, 0x00, 0x00, 0x00,
						0x44, 0x9a, 0x52, 0x2b,
					}))
				})
			})

			When("using a value of float32(-1234.5678)", func() {
				var err error
				var avp *diameter.AVP
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(100, 16777216, false, diameter.Float32, float32(-1234.5678))
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           100,
						VendorID:       16777216,
						VendorSpecific: true,
						Mandatory:      false,
						Length:         16,
						PaddedLength:   16,
						Data:           []byte{0xc4, 0x9a, 0x52, 0x2b},
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.Float32,
							TypedValue: float32(-1234.5678),
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0x00, 0x00, 0x64,
						0x80, 0x00, 0x00, 0x10,
						0x01, 0x00, 0x00, 0x00,
						0xc4, 0x9a, 0x52, 0x2b,
					}))
				})
			})

			When("using a value of float64(0)", func() {
				var err error
				BeforeEach(func() {
					_, err = diameter.NewTypedAVPErrorable(100, 16777216, false, diameter.Float32, float64(0))
				})

				It("returns an error", func() {
					Expect(err).ToNot(BeNil())
				})
			})

			When("using a value of int(65536)", func() {
				var err error
				var avp *diameter.AVP
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(100, 16777216, false, diameter.Float32, int(65536))
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           100,
						VendorID:       16777216,
						VendorSpecific: true,
						Mandatory:      false,
						Length:         16,
						PaddedLength:   16,
						Data:           []byte{0x47, 0x80, 0x00, 0x00},
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.Float32,
							TypedValue: float32(65536.0),
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0x00, 0x00, 0x64,
						0x80, 0x00, 0x00, 0x10,
						0x01, 0x00, 0x00, 0x00,
						0x47, 0x80, 0x00, 0x00,
					}))
				})
			})

			When("using a value of nil", func() {
				var err error
				BeforeEach(func() {
					_, err = diameter.NewTypedAVPErrorable(100, 16777216, false, diameter.Float32, nil)
				})

				It("returns an error", func() {
					Expect(err).ToNot(BeNil())
				})
			})

			When("using a value of '1.0' (string)", func() {
				var err error
				BeforeEach(func() {
					_, err = diameter.NewTypedAVPErrorable(100, 16777216, false, diameter.Float32, "1.0")
				})

				It("returns an error", func() {
					Expect(err).ToNot(BeNil())
				})
			})

		})

		Describe("creating custom AVP 16777215 (type Float64)", func() {
			When("using a value of float64(0)", func() {
				var err error
				var avp *diameter.AVP
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(16777215, 0, false, diameter.Float64, float64(0))
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           16777215,
						VendorID:       0,
						VendorSpecific: false,
						Mandatory:      false,
						Length:         16,
						PaddedLength:   16,
						Data:           []byte{0, 0, 0, 0, 0, 0, 0, 0},
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.Float64,
							TypedValue: float64(0),
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0xff, 0xff, 0xff,
						0x00, 0x00, 0x00, 0x10,
						0x00, 0x00, 0x00, 0x00,
						0x00, 0x00, 0x00, 0x00,
					}))
				})
			})

			When("using a value of float64(1234.5678)", func() {
				var err error
				var avp *diameter.AVP
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(16777215, 0, false, diameter.Float64, float64(1234.5678))
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           16777215,
						VendorID:       0,
						VendorSpecific: false,
						Mandatory:      false,
						Length:         16,
						PaddedLength:   16,
						Data:           []byte{0x40, 0x93, 0x4a, 0x45, 0x6d, 0x5c, 0xfa, 0xad},
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.Float64,
							TypedValue: float64(1234.5678),
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0xff, 0xff, 0xff,
						0x00, 0x00, 0x00, 0x10,
						0x40, 0x93, 0x4a, 0x45,
						0x6d, 0x5c, 0xfa, 0xad,
					}))
				})
			})

			When("using a value of float64(-1234.5678)", func() {
				var err error
				var avp *diameter.AVP
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(16777215, 0, false, diameter.Float64, float64(-1234.5678))
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           16777215,
						VendorID:       0,
						VendorSpecific: false,
						Mandatory:      false,
						Length:         16,
						PaddedLength:   16,
						Data:           []byte{0xc0, 0x93, 0x4a, 0x45, 0x6d, 0x5c, 0xfa, 0xad},
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.Float64,
							TypedValue: float64(-1234.5678),
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0xff, 0xff, 0xff,
						0x00, 0x00, 0x00, 0x10,
						0xc0, 0x93, 0x4a, 0x45,
						0x6d, 0x5c, 0xfa, 0xad,
					}))
				})
			})

			When("using a value of float64(9999999999999999999999999)", func() {
				var err error
				var avp *diameter.AVP
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(16777215, 0, false, diameter.Float64, float64(999999999999999999999999))
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           16777215,
						VendorID:       0,
						VendorSpecific: false,
						Mandatory:      false,
						Length:         16,
						PaddedLength:   16,
						Data:           []byte{0x44, 0xea, 0x78, 0x43, 0x79, 0xd9, 0x9d, 0xb4},
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.Float64,
							TypedValue: float64(999999999999999983222784.000000),
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0xff, 0xff, 0xff,
						0x00, 0x00, 0x00, 0x10,
						0x44, 0xea, 0x78, 0x43,
						0x79, 0xd9, 0x9d, 0xb4,
					}))
				})
			})

			When("using a value of int(65536)", func() {
				var err error
				var avp *diameter.AVP
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(16777215, 0, false, diameter.Float64, int(65536))
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           16777215,
						VendorID:       0,
						VendorSpecific: false,
						Mandatory:      false,
						Length:         16,
						PaddedLength:   16,
						Data:           []byte{0x40, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.Float64,
							TypedValue: float64(65536.0),
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0xff, 0xff, 0xff,
						0x00, 0x00, 0x00, 0x10,
						0x40, 0xf0, 0x00, 0x00,
						0x00, 0x00, 0x00, 0x00,
					}))
				})
			})

			When("using a value of nil", func() {
				var err error
				BeforeEach(func() {
					_, err = diameter.NewTypedAVPErrorable(16777215, 0, false, diameter.Float64, nil)
				})

				It("returns an error", func() {
					Expect(err).ToNot(BeNil())
				})
			})

			When("using a value of '1.0' (string)", func() {
				var err error
				BeforeEach(func() {
					_, err = diameter.NewTypedAVPErrorable(16777215, 0, false, diameter.Float64, "1.0")
				})

				It("returns an error", func() {
					Expect(err).ToNot(BeNil())
				})
			})

		})

		Describe("creating AVP Session-Id (type UTF8String)", func() {
			When("using a value of '' (the empty string)", func() {
				var err error
				var avp *diameter.AVP
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(263, 0, true, diameter.UTF8String, "")
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           263,
						VendorID:       0,
						VendorSpecific: false,
						Mandatory:      true,
						Length:         8,
						PaddedLength:   8,
						Data:           []byte{},
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.UTF8String,
							TypedValue: "",
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0x00, 0x01, 0x07,
						0x40, 0x00, 0x00, 0x08,
					}))
				})

			})

			When("using a value of 'accesspoint7.example.com;1876543210;523;mobile@200.1.1.88'", func() {
				var err error
				var avp *diameter.AVP
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(263, 0, true, diameter.UTF8String, "accesspoint7.example.com;1876543210;523;mobile@200.1.1.88")
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           263,
						VendorID:       0,
						VendorSpecific: false,
						Mandatory:      true,
						Length:         65,
						PaddedLength:   68,
						Data:           []byte("accesspoint7.example.com;1876543210;523;mobile@200.1.1.88"),
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.UTF8String,
							TypedValue: "accesspoint7.example.com;1876543210;523;mobile@200.1.1.88",
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0x00, 0x01, 0x07,
						0x40, 0x00, 0x00, 0x41,
						0x61, 0x63, 0x63, 0x65,
						0x73, 0x73, 0x70, 0x6f,
						0x69, 0x6e, 0x74, 0x37,
						0x2e, 0x65, 0x78, 0x61,
						0x6d, 0x70, 0x6c, 0x65,
						0x2e, 0x63, 0x6f, 0x6d,
						0x3b, 0x31, 0x38, 0x37,
						0x36, 0x35, 0x34, 0x33,
						0x32, 0x31, 0x30, 0x3b,
						0x35, 0x32, 0x33, 0x3b,
						0x6d, 0x6f, 0x62, 0x69,
						0x6c, 0x65, 0x40, 0x32,
						0x30, 0x30, 0x2e, 0x31,
						0x2e, 0x31, 0x2e, 0x38,
						0x38, 0x00, 0x00, 0x00,
					}))
				})
			})

			When("using a value of []byte('accesspoint7.example.com;1876543210;523;mobile@200.1.1.88')", func() {
				var err error
				var avp *diameter.AVP
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(263, 0, true, diameter.UTF8String, []byte("accesspoint7.example.com;1876543210;523;mobile@200.1.1.88"))
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           263,
						VendorID:       0,
						VendorSpecific: false,
						Mandatory:      true,
						Length:         65,
						PaddedLength:   68,
						Data:           []byte("accesspoint7.example.com;1876543210;523;mobile@200.1.1.88"),
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.UTF8String,
							TypedValue: "accesspoint7.example.com;1876543210;523;mobile@200.1.1.88",
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0x00, 0x01, 0x07,
						0x40, 0x00, 0x00, 0x41,
						0x61, 0x63, 0x63, 0x65,
						0x73, 0x73, 0x70, 0x6f,
						0x69, 0x6e, 0x74, 0x37,
						0x2e, 0x65, 0x78, 0x61,
						0x6d, 0x70, 0x6c, 0x65,
						0x2e, 0x63, 0x6f, 0x6d,
						0x3b, 0x31, 0x38, 0x37,
						0x36, 0x35, 0x34, 0x33,
						0x32, 0x31, 0x30, 0x3b,
						0x35, 0x32, 0x33, 0x3b,
						0x6d, 0x6f, 0x62, 0x69,
						0x6c, 0x65, 0x40, 0x32,
						0x30, 0x30, 0x2e, 0x31,
						0x2e, 0x31, 0x2e, 0x38,
						0x38, 0x00, 0x00, 0x00,
					}))
				})
			})

			When("using a value of 'ァアィイゥウェエォオカガキギクグケゲコゴサザ@example.com'", func() {
				var err error
				var avp *diameter.AVP
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(263, 0, true, diameter.UTF8String, "ァアィイゥウェエォオカガキギクグケゲコゴサザ@example.com")
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           263,
						VendorID:       0,
						VendorSpecific: false,
						Mandatory:      true,
						Length:         86,
						PaddedLength:   88,
						Data: []byte{0xe3, 0x82, 0xa1, 0xe3, 0x82, 0xa2, 0xe3, 0x82, 0xa3, 0xe3, 0x82, 0xa4, 0xe3, 0x82, 0xa5, 0xe3,
							0x82, 0xa6, 0xe3, 0x82, 0xa7, 0xe3, 0x82, 0xa8, 0xe3, 0x82, 0xa9, 0xe3, 0x82, 0xaa, 0xe3, 0x82,
							0xab, 0xe3, 0x82, 0xac, 0xe3, 0x82, 0xad, 0xe3, 0x82, 0xae, 0xe3, 0x82, 0xaf, 0xe3, 0x82, 0xb0,
							0xe3, 0x82, 0xb1, 0xe3, 0x82, 0xb2, 0xe3, 0x82, 0xb3, 0xe3, 0x82, 0xb4, 0xe3, 0x82, 0xb5, 0xe3,
							0x82, 0xb6, 0x40, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d},
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.UTF8String,
							TypedValue: "ァアィイゥウェエォオカガキギクグケゲコゴサザ@example.com",
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0x00, 0x01, 0x07,
						0x40, 0x00, 0x00, 0x56,
						0xe3, 0x82, 0xa1, 0xe3,
						0x82, 0xa2, 0xe3, 0x82,
						0xa3, 0xe3, 0x82, 0xa4,
						0xe3, 0x82, 0xa5, 0xe3,
						0x82, 0xa6, 0xe3, 0x82,
						0xa7, 0xe3, 0x82, 0xa8,
						0xe3, 0x82, 0xa9, 0xe3,
						0x82, 0xaa, 0xe3, 0x82,
						0xab, 0xe3, 0x82, 0xac,
						0xe3, 0x82, 0xad, 0xe3,
						0x82, 0xae, 0xe3, 0x82,
						0xaf, 0xe3, 0x82, 0xb0,
						0xe3, 0x82, 0xb1, 0xe3,
						0x82, 0xb2, 0xe3, 0x82,
						0xb3, 0xe3, 0x82, 0xb4,
						0xe3, 0x82, 0xb5, 0xe3,
						0x82, 0xb6, 0x40, 0x65,
						0x78, 0x61, 0x6d, 0x70,
						0x6c, 0x65, 0x2e, 0x63,
						0x6f, 0x6d, 0x00, 0x00,
					}))
				})
			})

			When("using a value of 'ァアィイゥウェエォオカガキギクグケゲコゴサザ@example.com' as []byte slice", func() {
				var err error
				var avp *diameter.AVP
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(
						263,
						0,
						true,
						diameter.UTF8String,
						[]byte{0xe3, 0x82, 0xa1, 0xe3, 0x82, 0xa2, 0xe3, 0x82, 0xa3, 0xe3, 0x82, 0xa4, 0xe3, 0x82, 0xa5, 0xe3,
							0x82, 0xa6, 0xe3, 0x82, 0xa7, 0xe3, 0x82, 0xa8, 0xe3, 0x82, 0xa9, 0xe3, 0x82, 0xaa, 0xe3, 0x82,
							0xab, 0xe3, 0x82, 0xac, 0xe3, 0x82, 0xad, 0xe3, 0x82, 0xae, 0xe3, 0x82, 0xaf, 0xe3, 0x82, 0xb0,
							0xe3, 0x82, 0xb1, 0xe3, 0x82, 0xb2, 0xe3, 0x82, 0xb3, 0xe3, 0x82, 0xb4, 0xe3, 0x82, 0xb5, 0xe3,
							0x82, 0xb6, 0x40, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d},
					)
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           263,
						VendorID:       0,
						VendorSpecific: false,
						Mandatory:      true,
						Length:         86,
						PaddedLength:   88,
						Data: []byte{0xe3, 0x82, 0xa1, 0xe3, 0x82, 0xa2, 0xe3, 0x82, 0xa3, 0xe3, 0x82, 0xa4, 0xe3, 0x82, 0xa5, 0xe3,
							0x82, 0xa6, 0xe3, 0x82, 0xa7, 0xe3, 0x82, 0xa8, 0xe3, 0x82, 0xa9, 0xe3, 0x82, 0xaa, 0xe3, 0x82,
							0xab, 0xe3, 0x82, 0xac, 0xe3, 0x82, 0xad, 0xe3, 0x82, 0xae, 0xe3, 0x82, 0xaf, 0xe3, 0x82, 0xb0,
							0xe3, 0x82, 0xb1, 0xe3, 0x82, 0xb2, 0xe3, 0x82, 0xb3, 0xe3, 0x82, 0xb4, 0xe3, 0x82, 0xb5, 0xe3,
							0x82, 0xb6, 0x40, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d},
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.UTF8String,
							TypedValue: "ァアィイゥウェエォオカガキギクグケゲコゴサザ@example.com",
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0x00, 0x01, 0x07,
						0x40, 0x00, 0x00, 0x56,
						0xe3, 0x82, 0xa1, 0xe3,
						0x82, 0xa2, 0xe3, 0x82,
						0xa3, 0xe3, 0x82, 0xa4,
						0xe3, 0x82, 0xa5, 0xe3,
						0x82, 0xa6, 0xe3, 0x82,
						0xa7, 0xe3, 0x82, 0xa8,
						0xe3, 0x82, 0xa9, 0xe3,
						0x82, 0xaa, 0xe3, 0x82,
						0xab, 0xe3, 0x82, 0xac,
						0xe3, 0x82, 0xad, 0xe3,
						0x82, 0xae, 0xe3, 0x82,
						0xaf, 0xe3, 0x82, 0xb0,
						0xe3, 0x82, 0xb1, 0xe3,
						0x82, 0xb2, 0xe3, 0x82,
						0xb3, 0xe3, 0x82, 0xb4,
						0xe3, 0x82, 0xb5, 0xe3,
						0x82, 0xb6, 0x40, 0x65,
						0x78, 0x61, 0x6d, 0x70,
						0x6c, 0x65, 0x2e, 0x63,
						0x6f, 0x6d, 0x00, 0x00,
					}))
				})

			})

			When("using a value of 'ァアィイゥウェエォオカガキギクグケゲコゴサザ@example.com' as []rune slice", func() {
				var err error
				var avp *diameter.AVP
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(
						263,
						0,
						true,
						diameter.UTF8String,
						[]rune{'ァ', 'ア', 'ィ', 'イ', 'ゥ', 'ウ', 'ェ', 'エ', 'ォ', 'オ', 'カ', 'ガ', 'キ', 'ギ', 'ク', 'グ', 'ケ', 'ゲ', 'コ', 'ゴ', 'サ', 'ザ', '@', 'e', 'x', 'a', 'm', 'p', 'l', 'e', '.', 'c', 'o', 'm'},
					)
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           263,
						VendorID:       0,
						VendorSpecific: false,
						Mandatory:      true,
						Length:         86,
						PaddedLength:   88,
						Data: []byte{0xe3, 0x82, 0xa1, 0xe3, 0x82, 0xa2, 0xe3, 0x82, 0xa3, 0xe3, 0x82, 0xa4, 0xe3, 0x82, 0xa5, 0xe3,
							0x82, 0xa6, 0xe3, 0x82, 0xa7, 0xe3, 0x82, 0xa8, 0xe3, 0x82, 0xa9, 0xe3, 0x82, 0xaa, 0xe3, 0x82,
							0xab, 0xe3, 0x82, 0xac, 0xe3, 0x82, 0xad, 0xe3, 0x82, 0xae, 0xe3, 0x82, 0xaf, 0xe3, 0x82, 0xb0,
							0xe3, 0x82, 0xb1, 0xe3, 0x82, 0xb2, 0xe3, 0x82, 0xb3, 0xe3, 0x82, 0xb4, 0xe3, 0x82, 0xb5, 0xe3,
							0x82, 0xb6, 0x40, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d},
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.UTF8String,
							TypedValue: "ァアィイゥウェエォオカガキギクグケゲコゴサザ@example.com",
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0x00, 0x01, 0x07,
						0x40, 0x00, 0x00, 0x56,
						0xe3, 0x82, 0xa1, 0xe3,
						0x82, 0xa2, 0xe3, 0x82,
						0xa3, 0xe3, 0x82, 0xa4,
						0xe3, 0x82, 0xa5, 0xe3,
						0x82, 0xa6, 0xe3, 0x82,
						0xa7, 0xe3, 0x82, 0xa8,
						0xe3, 0x82, 0xa9, 0xe3,
						0x82, 0xaa, 0xe3, 0x82,
						0xab, 0xe3, 0x82, 0xac,
						0xe3, 0x82, 0xad, 0xe3,
						0x82, 0xae, 0xe3, 0x82,
						0xaf, 0xe3, 0x82, 0xb0,
						0xe3, 0x82, 0xb1, 0xe3,
						0x82, 0xb2, 0xe3, 0x82,
						0xb3, 0xe3, 0x82, 0xb4,
						0xe3, 0x82, 0xb5, 0xe3,
						0x82, 0xb6, 0x40, 0x65,
						0x78, 0x61, 0x6d, 0x70,
						0x6c, 0x65, 0x2e, 0x63,
						0x6f, 0x6d, 0x00, 0x00,
					}))
				})

			})

			When("using a value of nil", func() {
				var err error
				BeforeEach(func() {
					_, err = diameter.NewTypedAVPErrorable(263, 0, true, diameter.UTF8String, nil)
				})

				It("returns an error", func() {
					Expect(err).ToNot(BeNil())
				})
			})

			When("using a value of 10", func() {
				var err error
				BeforeEach(func() {
					_, err = diameter.NewTypedAVPErrorable(263, 0, true, diameter.UTF8String, 10)
				})

				It("returns an error", func() {
					Expect(err).ToNot(BeNil())
				})
			})

			When("using a value of []byte{0xc3, 0x28} (not utf8)", func() {
				var err error
				BeforeEach(func() {
					_, err = diameter.NewTypedAVPErrorable(263, 0, true, diameter.UTF8String, []byte{0xc3, 0x28})
				})

				It("returns an error", func() {
					Expect(err).ToNot(BeNil())
				})
			})

			When("using a value of '\xc3\x28' (not utf8)", func() {
				var err error
				BeforeEach(func() {
					_, err = diameter.NewTypedAVPErrorable(263, 0, true, diameter.UTF8String, "\xc3\x28")
				})

				It("returns an error", func() {
					Expect(err).ToNot(BeNil())
				})
			})
		})

		Describe("creating AVP Charging-Rule-Name (type OctetString)", func() {
			When("using a value of []byte{}", func() {
				var err error
				var avp *diameter.AVP
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(1005, 10415, true, diameter.OctetString, []byte{})
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           1005,
						VendorID:       10415,
						VendorSpecific: true,
						Mandatory:      true,
						Length:         12,
						PaddedLength:   12,
						Data:           []byte{},
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.OctetString,
							TypedValue: []byte{},
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0x00, 0x03, 0xed,
						0xc0, 0x00, 0x00, 0x0c,
						0x00, 0x00, 0x28, 0xaf,
					}))
				})
			})

			When("using a value of []byte{0x00}", func() {
				var err error
				var avp *diameter.AVP
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(1005, 10415, true, diameter.OctetString, []byte{0x00})
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           1005,
						VendorID:       10415,
						VendorSpecific: true,
						Mandatory:      true,
						Length:         13,
						PaddedLength:   16,
						Data:           []byte{0x00},
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.OctetString,
							TypedValue: []byte{0x00},
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0x00, 0x03, 0xed,
						0xc0, 0x00, 0x00, 0x0d,
						0x00, 0x00, 0x28, 0xaf,
						0x00, 0x00, 0x00, 0x00,
					}))
				})
			})

			When("using a value of []byte{0x00, 0x01, 0x02, 0xde, 0xdf, 0xff}", func() {
				var err error
				var avp *diameter.AVP
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(1005, 10415, true, diameter.OctetString, []byte{0x00, 0x01, 0x02, 0xde, 0xdf, 0xff})
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           1005,
						VendorID:       10415,
						VendorSpecific: true,
						Mandatory:      true,
						Length:         18,
						PaddedLength:   20,
						Data:           []byte{0x00, 0x01, 0x02, 0xde, 0xdf, 0xff},
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.OctetString,
							TypedValue: []byte{0x00, 0x01, 0x02, 0xde, 0xdf, 0xff},
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0x00, 0x03, 0xed,
						0xc0, 0x00, 0x00, 0x12,
						0x00, 0x00, 0x28, 0xaf,
						0x00, 0x01, 0x02, 0xde,
						0xdf, 0xff, 0x00, 0x00,
					}))
				})
			})

			When("using a value of '' (the empty string)", func() {
				var err error
				var avp *diameter.AVP
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(1005, 10415, true, diameter.OctetString, "")
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           1005,
						VendorID:       10415,
						VendorSpecific: true,
						Mandatory:      true,
						Length:         12,
						PaddedLength:   12,
						Data:           []byte{},
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.OctetString,
							TypedValue: []byte{},
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0x00, 0x03, 0xed,
						0xc0, 0x00, 0x00, 0x0c,
						0x00, 0x00, 0x28, 0xaf,
					}))
				})
			})

			When("using a value of 'ァアィイゥウェエォオカガキギクグケゲコゴサザ@example.com'", func() {
				var err error
				var avp *diameter.AVP
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(1005, 10415, true, diameter.OctetString, "ァアィイゥウェエォオカガキギクグケゲコゴサザ@example.com")
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           1005,
						VendorID:       10415,
						VendorSpecific: true,
						Mandatory:      true,
						Length:         90,
						PaddedLength:   92,
						Data: []byte{0xe3, 0x82, 0xa1, 0xe3, 0x82, 0xa2, 0xe3, 0x82, 0xa3, 0xe3, 0x82, 0xa4, 0xe3, 0x82, 0xa5, 0xe3,
							0x82, 0xa6, 0xe3, 0x82, 0xa7, 0xe3, 0x82, 0xa8, 0xe3, 0x82, 0xa9, 0xe3, 0x82, 0xaa, 0xe3, 0x82,
							0xab, 0xe3, 0x82, 0xac, 0xe3, 0x82, 0xad, 0xe3, 0x82, 0xae, 0xe3, 0x82, 0xaf, 0xe3, 0x82, 0xb0,
							0xe3, 0x82, 0xb1, 0xe3, 0x82, 0xb2, 0xe3, 0x82, 0xb3, 0xe3, 0x82, 0xb4, 0xe3, 0x82, 0xb5, 0xe3,
							0x82, 0xb6, 0x40, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d},
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:     "",
							DataType: diameter.OctetString,
							TypedValue: []byte{0xe3, 0x82, 0xa1, 0xe3, 0x82, 0xa2, 0xe3, 0x82, 0xa3, 0xe3, 0x82, 0xa4, 0xe3, 0x82, 0xa5, 0xe3,
								0x82, 0xa6, 0xe3, 0x82, 0xa7, 0xe3, 0x82, 0xa8, 0xe3, 0x82, 0xa9, 0xe3, 0x82, 0xaa, 0xe3, 0x82,
								0xab, 0xe3, 0x82, 0xac, 0xe3, 0x82, 0xad, 0xe3, 0x82, 0xae, 0xe3, 0x82, 0xaf, 0xe3, 0x82, 0xb0,
								0xe3, 0x82, 0xb1, 0xe3, 0x82, 0xb2, 0xe3, 0x82, 0xb3, 0xe3, 0x82, 0xb4, 0xe3, 0x82, 0xb5, 0xe3,
								0x82, 0xb6, 0x40, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d},
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0x00, 0x03, 0xed,
						0xc0, 0x00, 0x00, 0x5a,
						0x00, 0x00, 0x28, 0xaf,
						0xe3, 0x82, 0xa1, 0xe3,
						0x82, 0xa2, 0xe3, 0x82,
						0xa3, 0xe3, 0x82, 0xa4,
						0xe3, 0x82, 0xa5, 0xe3,
						0x82, 0xa6, 0xe3, 0x82,
						0xa7, 0xe3, 0x82, 0xa8,
						0xe3, 0x82, 0xa9, 0xe3,
						0x82, 0xaa, 0xe3, 0x82,
						0xab, 0xe3, 0x82, 0xac,
						0xe3, 0x82, 0xad, 0xe3,
						0x82, 0xae, 0xe3, 0x82,
						0xaf, 0xe3, 0x82, 0xb0,
						0xe3, 0x82, 0xb1, 0xe3,
						0x82, 0xb2, 0xe3, 0x82,
						0xb3, 0xe3, 0x82, 0xb4,
						0xe3, 0x82, 0xb5, 0xe3,
						0x82, 0xb6, 0x40, 0x65,
						0x78, 0x61, 0x6d, 0x70,
						0x6c, 0x65, 0x2e, 0x63,
						0x6f, 0x6d, 0x00, 0x00,
					}))
				})
			})

			When("using a value of []byte{0xc3, 0x28} (not utf8)", func() {
				var err error
				var avp *diameter.AVP
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(1005, 10415, true, diameter.OctetString, []byte{0xc3, 0x28})
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           1005,
						VendorID:       10415,
						VendorSpecific: true,
						Mandatory:      true,
						Length:         14,
						PaddedLength:   16,
						Data:           []byte{0xc3, 0x28},
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.OctetString,
							TypedValue: []byte{0xc3, 0x28},
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0x00, 0x03, 0xed,
						0xc0, 0x00, 0x00, 0x0e,
						0x00, 0x00, 0x28, 0xaf,
						0xc3, 0x28, 0x00, 0x00,
					}))
				})
			})

			When("using a value of '\xc3\x28' (string that is not utf8)", func() {
				var err error
				var avp *diameter.AVP
				BeforeEach(func() {
					avp, err = diameter.NewTypedAVPErrorable(1005, 10415, true, diameter.OctetString, "\xc3\x28")
				})

				It("does not return an error", func() {
					Expect(err).To(BeNil())
				})

				It("properly sets AVP exported fields", func() {
					Expect(avp).To(Equal(&diameter.AVP{
						Code:           1005,
						VendorID:       10415,
						VendorSpecific: true,
						Mandatory:      true,
						Length:         14,
						PaddedLength:   16,
						Data:           []byte{0xc3, 0x28},
						ExtendedAttributes: &diameter.AVPExtendedAttributes{
							Name:       "",
							DataType:   diameter.OctetString,
							TypedValue: []byte{0xc3, 0x28},
						},
					}))
				})

				It("properly Encodes", func() {
					Expect(avp.Encode()).To(Equal([]byte{
						0x00, 0x00, 0x03, 0xed,
						0xc0, 0x00, 0x00, 0x0e,
						0x00, 0x00, 0x28, 0xaf,
						0xc3, 0x28, 0x00, 0x00,
					}))
				})
			})

			When("using a value of nil", func() {
				var err error
				BeforeEach(func() {
					_, err = diameter.NewTypedAVPErrorable(1005, 10415, true, diameter.OctetString, nil)
				})

				It("returns an error", func() {
					Expect(err).ToNot(BeNil())
				})
			})

			When("using a value of 10", func() {
				var err error
				BeforeEach(func() {
					_, err = diameter.NewTypedAVPErrorable(1005, 10415, true, diameter.OctetString, 10)
				})

				It("returns an error", func() {
					Expect(err).ToNot(BeNil())
				})
			})
		})
	})

	Describe("creating AVP Event-Timestamp (type Time)", func() {
		When("using a value of time.Unix(1717298560, 0)", func() {
			var err error
			var avp *diameter.AVP
			t := time.Unix(1717298560, 0)
			BeforeEach(func() {
				avp, err = diameter.NewTypedAVPErrorable(55, 0, true, diameter.Time, t)
			})

			It("does not return an error", func() {
				Expect(err).To(BeNil())
			})

			It("properly sets AVP exported fields", func() {
				Expect(avp).To(Equal(&diameter.AVP{
					Code:           55,
					VendorID:       0,
					VendorSpecific: false,
					Mandatory:      true,
					Length:         12,
					PaddedLength:   12,
					Data:           []byte{0xea, 0x06, 0x64, 0x00},
					ExtendedAttributes: &diameter.AVPExtendedAttributes{
						Name:       "",
						DataType:   diameter.Time,
						TypedValue: &t,
					},
				}))
			})

			It("properly Encodes", func() {
				Expect(avp.Encode()).To(Equal([]byte{
					0x00, 0x00, 0x00, 0x37,
					0x40, 0x00, 0x00, 0x0c,
					0xea, 0x06, 0x64, 0x00,
				}))
			})
		})

		When("using a value of *time.Unix(1717298560, 0)", func() {
			var err error
			var avp *diameter.AVP
			t := time.Unix(1717298560, 0)
			BeforeEach(func() {
				avp, err = diameter.NewTypedAVPErrorable(55, 0, true, diameter.Time, &t)
			})

			It("does not return an error", func() {
				Expect(err).To(BeNil())
			})

			It("properly sets AVP exported fields", func() {
				Expect(avp).To(Equal(&diameter.AVP{
					Code:           55,
					VendorID:       0,
					VendorSpecific: false,
					Mandatory:      true,
					Length:         12,
					PaddedLength:   12,
					Data:           []byte{0xea, 0x06, 0x64, 0x00},
					ExtendedAttributes: &diameter.AVPExtendedAttributes{
						Name:       "",
						DataType:   diameter.Time,
						TypedValue: &t,
					},
				}))
			})

			It("properly Encodes", func() {
				Expect(avp.Encode()).To(Equal([]byte{
					0x00, 0x00, 0x00, 0x37,
					0x40, 0x00, 0x00, 0x0c,
					0xea, 0x06, 0x64, 0x00,
				}))
			})
		})

		When("using a value of int(3926287360)", func() {
			var err error
			var avp *diameter.AVP

			t := time.Unix(1717298560, 0)
			BeforeEach(func() {
				avp, err = diameter.NewTypedAVPErrorable(55, 0, true, diameter.Time, int(3926287360))
			})

			It("does not return an error", func() {
				Expect(err).To(BeNil())
			})

			It("properly sets AVP exported fields", func() {
				// since the returned value is a pointer, normal Equal() won't work, so just compare
				// produced value
				Expect(avp.ExtendedAttributes.TypedValue.(*time.Time).Unix()).To(Equal(t.Unix()))
			})

			It("properly Encodes", func() {
				Expect(avp.Encode()).To(Equal([]byte{
					0x00, 0x00, 0x00, 0x37,
					0x40, 0x00, 0x00, 0x0c,
					0xea, 0x06, 0x64, 0x00,
				}))
			})
		})

		When("using a byte slice for the value", func() {
			var err error
			var avp *diameter.AVP

			t := time.Unix(1717298560, 0)
			BeforeEach(func() {
				avp, err = diameter.NewTypedAVPErrorable(55, 0, true, diameter.Time, []byte{0xea, 0x06, 0x64, 0x00})
			})

			It("does not return an error", func() {
				Expect(err).To(BeNil())
			})

			It("properly sets AVP exported fields", func() {
				// since the returned value is a pointer, normal Equal() won't work, so just compare
				// produced value
				Expect(avp.ExtendedAttributes.TypedValue.(*time.Time).Unix()).To(Equal(t.Unix()))
			})

			It("properly Encodes", func() {
				Expect(avp.Encode()).To(Equal([]byte{
					0x00, 0x00, 0x00, 0x37,
					0x40, 0x00, 0x00, 0x0c,
					0xea, 0x06, 0x64, 0x00,
				}))
			})
		})

		When("using a byte slice of size 0 for the value", func() {
			var err error

			BeforeEach(func() {
				_, err = diameter.NewTypedAVPErrorable(55, 0, true, diameter.Time, []byte{})
			})

			It("returns an error", func() {
				Expect(err).ToNot(BeNil())
			})
		})

		When("using a byte slice of size 5 for the value", func() {
			var err error

			BeforeEach(func() {
				_, err = diameter.NewTypedAVPErrorable(55, 0, true, diameter.Time, []byte{0xea, 0x06, 0x64, 0x00, 0x00})
			})

			It("returns an error", func() {
				Expect(err).ToNot(BeNil())
			})
		})

		When("using a negative int for a value", func() {
			var err error

			BeforeEach(func() {
				_, err = diameter.NewTypedAVPErrorable(55, 0, true, diameter.Time, int(-1))
			})

			It("returns an error", func() {
				Expect(err).ToNot(BeNil())
			})
		})
	})

	Describe("creating AVP Host-IP-Address (type Address)", func() {
		When("using a valid, IPv4-based *diameter.AddressType", func() {
			var err error
			var avp *diameter.AVP

			a := diameter.NewAddressTypeFromIP(net.ParseIP("10.254.10.1"))
			BeforeEach(func() {
				avp, err = diameter.NewTypedAVPErrorable(257, 0, true, diameter.Address, a)
			})

			It("does not return an error", func() {
				Expect(err).To(BeNil())
			})

			It("properly sets AVP exported fields", func() {
				Expect(avp).To(Equal(&diameter.AVP{
					Code:           257,
					VendorID:       0,
					VendorSpecific: false,
					Mandatory:      true,
					Length:         14,
					PaddedLength:   16,
					Data:           []byte{0x00, 0x01, 10, 254, 10, 1},
					ExtendedAttributes: &diameter.AVPExtendedAttributes{
						Name:       "",
						DataType:   diameter.Address,
						TypedValue: diameter.AddressType([]byte{0x00, 0x01, 10, 254, 10, 1}),
					},
				}))
			})

			It("properly Encodes", func() {
				Expect(avp.Encode()).To(Equal([]byte{
					0x00, 0x00, 0x01, 0x01,
					0x40, 0x00, 0x00, 0x0e,
					0x00, 0x01, 0x0a, 0xfe,
					0x0a, 0x01, 0x00, 0x00,
				}))
			})
		})

		When("using a valid, IPv6-based *diameter.AddressType", func() {
			var err error
			var avp *diameter.AVP

			a := diameter.NewAddressTypeFromIP(net.ParseIP("fd00:abcd:0:1::1"))
			BeforeEach(func() {
				avp, err = diameter.NewTypedAVPErrorable(257, 0, true, diameter.Address, a)
			})

			It("does not return an error", func() {
				Expect(err).To(BeNil())
			})

			It("properly sets AVP exported fields", func() {
				Expect(avp).To(Equal(&diameter.AVP{
					Code:           257,
					VendorID:       0,
					VendorSpecific: false,
					Mandatory:      true,
					Length:         26,
					PaddedLength:   28,
					Data:           []byte{0x00, 0x02, 0xfd, 0x00, 0xab, 0xcd, 0x00, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0, 0, 0x01},
					ExtendedAttributes: &diameter.AVPExtendedAttributes{
						Name:       "",
						DataType:   diameter.Address,
						TypedValue: diameter.AddressType([]byte{0x00, 0x02, 0xfd, 0x00, 0xab, 0xcd, 0x00, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0, 0, 0x01}),
					},
				}))
			})

			It("properly Encodes", func() {
				Expect(avp.Encode()).To(Equal([]byte{
					0x00, 0x00, 0x01, 0x01,
					0x40, 0x00, 0x00, 0x1a,
					0x00, 0x02, 0xfd, 0x00,
					0xab, 0xcd, 0x00, 0x00,
					0x00, 0x01, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00,
					0x00, 0x01, 0x00, 0x00,
				}))
			})
		})

		When("using a valid, IPv4-based diameter.AddressType", func() {
			var err error
			var avp *diameter.AVP

			a := diameter.NewAddressTypeFromIP(net.ParseIP("10.254.10.1"))
			BeforeEach(func() {
				avp, err = diameter.NewTypedAVPErrorable(257, 0, true, diameter.Address, a)
			})

			It("does not return an error", func() {
				Expect(err).To(BeNil())
			})

			It("properly sets AVP exported fields", func() {
				Expect(avp).To(Equal(&diameter.AVP{
					Code:           257,
					VendorID:       0,
					VendorSpecific: false,
					Mandatory:      true,
					Length:         14,
					PaddedLength:   16,
					Data:           []byte{0x00, 0x01, 10, 254, 10, 1},
					ExtendedAttributes: &diameter.AVPExtendedAttributes{
						Name:       "",
						DataType:   diameter.Address,
						TypedValue: diameter.AddressType([]byte{0x00, 0x01, 10, 254, 10, 1}),
					},
				}))
			})

			It("properly Encodes", func() {
				Expect(avp.Encode()).To(Equal([]byte{
					0x00, 0x00, 0x01, 0x01,
					0x40, 0x00, 0x00, 0x0e,
					0x00, 0x01, 0x0a, 0xfe,
					0x0a, 0x01, 0x00, 0x00,
				}))
			})
		})

		When("using a valid IPv4 net.IP", func() {
			var err error
			var avp *diameter.AVP

			BeforeEach(func() {
				avp, err = diameter.NewTypedAVPErrorable(257, 0, true, diameter.Address, net.ParseIP("0.0.0.0"))
			})

			It("does not return an error", func() {
				Expect(err).To(BeNil())
			})

			It("properly sets AVP exported fields", func() {
				Expect(avp).To(Equal(&diameter.AVP{
					Code:           257,
					VendorID:       0,
					VendorSpecific: false,
					Mandatory:      true,
					Length:         14,
					PaddedLength:   16,
					Data:           []byte{0x00, 0x01, 0, 0, 0, 0},
					ExtendedAttributes: &diameter.AVPExtendedAttributes{
						Name:       "",
						DataType:   diameter.Address,
						TypedValue: diameter.AddressType([]byte{0x00, 0x01, 0, 0, 0, 0}),
					},
				}))
			})

			It("properly Encodes", func() {
				Expect(avp.Encode()).To(Equal([]byte{
					0x00, 0x00, 0x01, 0x01,
					0x40, 0x00, 0x00, 0x0e,
					0x00, 0x01, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00,
				}))
			})
		})

		When("using a valid IPv6 *net.IP", func() {
			var err error
			var avp *diameter.AVP

			n := net.ParseIP("::")
			BeforeEach(func() {
				avp, err = diameter.NewTypedAVPErrorable(257, 0, true, diameter.Address, &n)
			})

			It("does not return an error", func() {
				Expect(err).To(BeNil())
			})

			It("properly sets AVP exported fields", func() {
				Expect(avp).To(Equal(&diameter.AVP{
					Code:           257,
					VendorID:       0,
					VendorSpecific: false,
					Mandatory:      true,
					Length:         26,
					PaddedLength:   28,
					Data:           []byte{0x00, 0x02, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
					ExtendedAttributes: &diameter.AVPExtendedAttributes{
						Name:       "",
						DataType:   diameter.Address,
						TypedValue: diameter.AddressType([]byte{0x00, 0x02, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}),
					},
				}))
			})

			It("properly Encodes", func() {
				Expect(avp.Encode()).To(Equal([]byte{
					0x00, 0x00, 0x01, 0x01,
					0x40, 0x00, 0x00, 0x1a,
					0x00, 0x02, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00,
				}))
			})
		})

		When("using a valid IPv6 net.IPAddr", func() {
			var err error
			var avp *diameter.AVP

			n, _ := net.ResolveIPAddr("ip", "fd00:abcd:0:1::1")
			BeforeEach(func() {
				avp, err = diameter.NewTypedAVPErrorable(257, 0, true, diameter.Address, *n)
			})

			It("does not return an error", func() {
				Expect(err).To(BeNil())
			})

			It("properly sets AVP exported fields", func() {
				Expect(avp).To(Equal(&diameter.AVP{
					Code:           257,
					VendorID:       0,
					VendorSpecific: false,
					Mandatory:      true,
					Length:         26,
					PaddedLength:   28,
					Data:           []byte{0x00, 0x02, 0xfd, 0x00, 0xab, 0xcd, 0x00, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0, 0x00, 0x01},
					ExtendedAttributes: &diameter.AVPExtendedAttributes{
						Name:       "",
						DataType:   diameter.Address,
						TypedValue: diameter.AddressType([]byte{0x00, 0x02, 0xfd, 0x00, 0xab, 0xcd, 0x00, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0, 0x00, 0x01}),
					},
				}))
			})

			It("properly Encodes", func() {
				Expect(avp.Encode()).To(Equal([]byte{
					0x00, 0x00, 0x01, 0x01,
					0x40, 0x00, 0x00, 0x1a,
					0x00, 0x02, 0xfd, 0x00,
					0xab, 0xcd, 0x00, 0x00,
					0x00, 0x01, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00,
					0x00, 0x01, 0x00, 0x00,
				}))
			})
		})

		When("using a valid IPv4 *net.IPAddr", func() {
			var err error
			var avp *diameter.AVP

			n, _ := net.ResolveIPAddr("ip", "255.255.255.255")
			BeforeEach(func() {
				avp, err = diameter.NewTypedAVPErrorable(257, 0, true, diameter.Address, n)
			})

			It("does not return an error", func() {
				Expect(err).To(BeNil())
			})

			It("properly sets AVP exported fields", func() {
				Expect(avp).To(Equal(&diameter.AVP{
					Code:           257,
					VendorID:       0,
					VendorSpecific: false,
					Mandatory:      true,
					Length:         14,
					PaddedLength:   16,
					Data:           []byte{0x00, 0x01, 255, 255, 255, 255},
					ExtendedAttributes: &diameter.AVPExtendedAttributes{
						Name:       "",
						DataType:   diameter.Address,
						TypedValue: diameter.AddressType([]byte{0x00, 0x01, 255, 255, 255, 255}),
					},
				}))
			})

			It("properly Encodes", func() {
				Expect(avp.Encode()).To(Equal([]byte{
					0x00, 0x00, 0x01, 0x01,
					0x40, 0x00, 0x00, 0x0e,
					0x00, 0x01, 0xff, 0xff,
					0xff, 0xff, 0x00, 0x00,
				}))
			})
		})

		When("using an AddressType with AddressFamilyNumber MAC48Bit", func() {
			var err error
			var avp *diameter.AVP

			a := diameter.NewAddressType(diameter.MAC48Bit, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06})
			BeforeEach(func() {
				avp, err = diameter.NewTypedAVPErrorable(257, 0, true, diameter.Address, a)
			})

			It("does not return an error", func() {
				Expect(err).To(BeNil())
			})

			It("properly sets AVP exported fields", func() {
				Expect(avp).To(Equal(&diameter.AVP{
					Code:           257,
					VendorID:       0,
					VendorSpecific: false,
					Mandatory:      true,
					Length:         17,
					PaddedLength:   20,
					Data:           []byte{0x40, 0x05, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
					ExtendedAttributes: &diameter.AVPExtendedAttributes{
						Name:       "",
						DataType:   diameter.Address,
						TypedValue: diameter.AddressType([]byte{0x40, 0x05, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06}),
					},
				}))
			})

			It("properly Encodes", func() {
				Expect(avp.Encode()).To(Equal([]byte{
					0x00, 0x00, 0x01, 0x01,
					0x40, 0x00, 0x00, 0x11,
					0x40, 0x05, 0x00, 0x01,
					0x02, 0x03, 0x04, 0x05,
					0x06, 0x00, 0x00, 0x00,
				}))
			})
		})

		When("using a value of nil", func() {
			var err error
			BeforeEach(func() {
				_, err = diameter.NewTypedAVPErrorable(257, 0, true, diameter.Address, nil)
			})

			It("returns an error", func() {
				Expect(err).ToNot(BeNil())
			})
		})

		When("using a value of 10", func() {
			var err error
			BeforeEach(func() {
				_, err = diameter.NewTypedAVPErrorable(257, 0, true, diameter.Address, 10)
			})

			It("returns an error", func() {
				Expect(err).ToNot(BeNil())
			})
		})

		When("using a value of '10.10.10.10' (a string)", func() {
			var err error
			BeforeEach(func() {
				_, err = diameter.NewTypedAVPErrorable(257, 0, true, diameter.Address, "10.10.10.10")
			})

			It("returns an error", func() {
				Expect(err).ToNot(BeNil())
			})
		})

		When("using a byte slice value", func() {
			var err error
			BeforeEach(func() {
				_, err = diameter.NewTypedAVPErrorable(257, 0, true, diameter.Address, []byte{0, 1, 10, 10, 10, 10})
			})

			It("returns an error", func() {
				Expect(err).ToNot(BeNil())
			})
		})

	})

	Describe("creating AVP Packet-Filter-Content (type IPFilterRule)", func() {
		When("using the value 'permit in ip from 0.0.0.0/0 to 10.10.10.0/24' (a string)", func() {
			var err error
			var avp *diameter.AVP

			BeforeEach(func() {
				avp, err = diameter.NewTypedAVPErrorable(1059, 10415, true, diameter.IPFilterRule, "permit in ip from 0.0.0.0/0 to 10.10.10.0/24")
			})

			It("does not return an error", func() {
				Expect(err).To(BeNil())
			})

			It("properly sets AVP exported fields", func() {
				Expect(avp).To(Equal(&diameter.AVP{
					Code:           1059,
					VendorID:       10415,
					VendorSpecific: true,
					Mandatory:      true,
					Length:         56,
					PaddedLength:   56,
					Data:           []byte("permit in ip from 0.0.0.0/0 to 10.10.10.0/24"),
					ExtendedAttributes: &diameter.AVPExtendedAttributes{
						Name:       "",
						DataType:   diameter.IPFilterRule,
						TypedValue: "permit in ip from 0.0.0.0/0 to 10.10.10.0/24",
					},
				}))
			})

			It("properly Encodes", func() {
				Expect(avp.Encode()).To(Equal([]byte{
					0x00, 0x00, 0x04, 0x23,
					0xc0, 0x00, 0x00, 0x38,
					0x00, 0x00, 0x28, 0xaf,
					0x70, 0x65, 0x72, 0x6d,
					0x69, 0x74, 0x20, 0x69,
					0x6e, 0x20, 0x69, 0x70,
					0x20, 0x66, 0x72, 0x6f,
					0x6d, 0x20, 0x30, 0x2e,
					0x30, 0x2e, 0x30, 0x2e,
					0x30, 0x2f, 0x30, 0x20,
					0x74, 0x6f, 0x20, 0x31,
					0x30, 0x2e, 0x31, 0x30,
					0x2e, 0x31, 0x30, 0x2e,
					0x30, 0x2f, 0x32, 0x34,
				}))
			})
		})

		When("using the value 'permit in ip from 0.0.0.0/0 to 10.10.10.0/24' as a byte slice", func() {
			var err error
			var avp *diameter.AVP

			BeforeEach(func() {
				avp, err = diameter.NewTypedAVPErrorable(1059, 10415, true, diameter.IPFilterRule, []byte("permit in ip from 0.0.0.0/0 to 10.10.10.0/24"))
			})

			It("does not return an error", func() {
				Expect(err).To(BeNil())
			})

			It("properly sets AVP exported fields", func() {
				Expect(avp).To(Equal(&diameter.AVP{
					Code:           1059,
					VendorID:       10415,
					VendorSpecific: true,
					Mandatory:      true,
					Length:         56,
					PaddedLength:   56,
					Data:           []byte("permit in ip from 0.0.0.0/0 to 10.10.10.0/24"),
					ExtendedAttributes: &diameter.AVPExtendedAttributes{
						Name:       "",
						DataType:   diameter.IPFilterRule,
						TypedValue: "permit in ip from 0.0.0.0/0 to 10.10.10.0/24",
					},
				}))
			})

			It("properly Encodes", func() {
				Expect(avp.Encode()).To(Equal([]byte{
					0x00, 0x00, 0x04, 0x23,
					0xc0, 0x00, 0x00, 0x38,
					0x00, 0x00, 0x28, 0xaf,
					0x70, 0x65, 0x72, 0x6d,
					0x69, 0x74, 0x20, 0x69,
					0x6e, 0x20, 0x69, 0x70,
					0x20, 0x66, 0x72, 0x6f,
					0x6d, 0x20, 0x30, 0x2e,
					0x30, 0x2e, 0x30, 0x2e,
					0x30, 0x2f, 0x30, 0x20,
					0x74, 0x6f, 0x20, 0x31,
					0x30, 0x2e, 0x31, 0x30,
					0x2e, 0x31, 0x30, 0x2e,
					0x30, 0x2f, 0x32, 0x34,
				}))
			})
		})

		When("using a the value nil", func() {
			var err error

			BeforeEach(func() {
				_, err = diameter.NewTypedAVPErrorable(1059, 10415, true, diameter.IPFilterRule, nil)
			})

			It("returns an error", func() {
				Expect(err).ToNot(BeNil())
			})
		})

		When("using a the value 10", func() {
			var err error

			BeforeEach(func() {
				_, err = diameter.NewTypedAVPErrorable(1059, 10415, true, diameter.IPFilterRule, 10)
			})

			It("returns an error", func() {
				Expect(err).ToNot(BeNil())
			})
		})
	})

	Describe("creating AVP Vendor-Specific-Applicaiton-Id (type Grouped)", func() {
		When("using as a value an empty AVP set", func() {
			var err error
			var avp *diameter.AVP

			BeforeEach(func() {
				avp, err = diameter.NewTypedAVPErrorable(260, 0, true, diameter.Grouped, []*diameter.AVP{})
			})

			It("does not return an error", func() {
				Expect(err).To(BeNil())
			})

			It("properly sets AVP exported fields", func() {
				Expect(avp).To(Equal(&diameter.AVP{
					Code:           260,
					VendorID:       0,
					VendorSpecific: false,
					Mandatory:      true,
					Length:         8,
					PaddedLength:   8,
					Data:           []byte{},
					ExtendedAttributes: &diameter.AVPExtendedAttributes{
						Name:       "",
						DataType:   diameter.Grouped,
						TypedValue: []*diameter.AVP{},
					},
				}))
			})

			It("properly Encodes", func() {
				Expect(avp.Encode()).To(Equal([]byte{
					0x00, 0x00, 0x01, 0x04,
					0x40, 0x00, 0x00, 0x08,
				}))
			})
		})

		When("using as a value an AVP set", func() {
			var err error
			var avp *diameter.AVP

			groupedAvps := []*diameter.AVP{
				diameter.NewTypedAVP(266, 0, true, diameter.Unsigned32, 10145),
				diameter.NewTypedAVP(258, 0, true, diameter.Unsigned32, 100),
			}

			BeforeEach(func() {
				avp, err = diameter.NewTypedAVPErrorable(260, 0, true, diameter.Grouped, groupedAvps)
			})

			It("does not return an error", func() {
				Expect(err).To(BeNil())
			})

			It("properly sets AVP exported fields", func() {
				Expect(avp).To(Equal(&diameter.AVP{
					Code:           260,
					VendorID:       0,
					VendorSpecific: false,
					Mandatory:      true,
					Length:         32,
					PaddedLength:   32,
					Data:           []byte{0x0, 0x0, 0x01, 0x0a, 0x40, 0x00, 0x00, 0xc, 0x0, 0x0, 0x27, 0xa1, 0x0, 0x0, 0x01, 0x02, 0x40, 0x0, 0x0, 0x0c, 0x0, 0x0, 0x0, 0x64},
					ExtendedAttributes: &diameter.AVPExtendedAttributes{
						Name:       "",
						DataType:   diameter.Grouped,
						TypedValue: groupedAvps,
					},
				}))
			})

			It("properly Encodes", func() {
				Expect(avp.Encode()).To(Equal([]byte{
					0x00, 0x00, 0x01, 0x04,
					0x40, 0x00, 0x00, 0x20,
					0x00, 0x00, 0x01, 0x0a,
					0x40, 0x00, 0x00, 0x0c,
					0x00, 0x00, 0x27, 0xa1,
					0x00, 0x00, 0x01, 0x02,
					0x40, 0x00, 0x00, 0x0c,
					0x00, 0x00, 0x00, 0x64,
				}))
			})

		})

		When("using as a byte slice as a value", func() {
			var err error

			BeforeEach(func() {
				_, err = diameter.NewTypedAVPErrorable(260, 0, true, diameter.Grouped, []byte{
					0x0, 0x0, 0x01, 0x0a, 0x40, 0x00, 0x00, 0xc, 0x0, 0x0, 0x27, 0xa1, 0x0, 0x0, 0x01, 0x02, 0x40, 0x0, 0x0, 0x0c, 0x0, 0x0, 0x0, 0x64,
				})
			})

			It("returns an error", func() {
				Expect(err).ToNot(BeNil())
			})
		})

		When("using a nil value", func() {
			var err error

			BeforeEach(func() {
				_, err = diameter.NewTypedAVPErrorable(260, 0, true, diameter.Grouped, nil)
			})

			It("returns an error", func() {
				Expect(err).ToNot(BeNil())
			})
		})
	})

	Describe("creating an AVP with an invalid type", func() {
		_, err := diameter.NewTypedAVPErrorable(100, 100, true, diameter.AVPDataType(0xfefefefe), []byte{})

		It("returns an error", func() {
			Expect(err).ToNot(BeNil())
		})
	})
})
