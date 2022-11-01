package avcodec

//#include <libavutil/avutil.h>
//#include <libavcodec/avcodec.h>
//
//#ifdef CODEC_CAP_HWACCEL
//#define GO_CODEC_CAP_HWACCEL CODEC_CAP_HWACCEL
//#else
//#define GO_CODEC_CAP_HWACCEL 0
//#endif
//
//#ifdef CODEC_CAP_HWACCEL_VDPAU
//#define GO_CODEC_CAP_HWACCEL_VDPAU CODEC_CAP_HWACCEL_VDPAU
//#else
//#define GO_CODEC_CAP_HWACCEL_VDPAU 0
//#endif
//
//#ifdef FF_DCT_INT
//#define GO_FF_DCT_INT FF_DCT_INT
//#else
//#define GO_FF_DCT_INT 0
//#endif
//
//#ifdef FF_IDCT_SH4
//#define GO_FF_IDCT_SH4 FF_IDCT_SH4
//#else
//#define GO_FF_IDCT_SH4 0
//#endif
//
//#ifdef FF_IDCT_IPP
//#define GO_FF_IDCT_IPP FF_IDCT_IPP
//#else
//#define GO_FF_IDCT_IPP 0
//#endif
//
//#ifdef FF_IDCT_XVIDMMX
//#define GO_FF_IDCT_XVIDMMX FF_IDCT_XVIDMMX
//#else
//#define GO_FF_IDCT_XVIDMMX 0
//#endif
//
//#ifdef FF_IDCT_SIMPLEVIS
//#define GO_FF_IDCT_SIMPLEVIS FF_IDCT_SIMPLEVIS
//#else
//#define GO_FF_IDCT_SIMPLEVIS 0
//#endif
//
//#ifdef FF_IDCT_SIMPLEALPHA
//#define GO_FF_IDCT_SIMPLEALPHA FF_IDCT_SIMPLEALPHA
//#else
//#define GO_FF_IDCT_SIMPLEALPHA 0
//#endif
//
//static const AVPacketSideData *go_av_packetsidedata_get(const AVPacketSideData *side_data, int n)
//{
//  return &side_data[n];
//}
//
//static const AVRational *go_av_rational_get(const AVRational *r, int n)
//{
//  if (r == NULL)
//  {
//    return NULL;
//  }
//  return &r[n];
//}
//
//static enum AVPixelFormat *go_av_pixfmt_get(enum AVPixelFormat *pixfmt, int n)
//{
//  if (pixfmt == NULL)
//  {
//    return NULL;
//  }
//  return &pixfmt[n];
//}
//
//static enum AVSampleFormat *go_av_samplefmt_get(enum AVSampleFormat *samplefmt, int n)
//{
//  if (samplefmt == NULL)
//  {
//    return NULL;
//  }
//  return &samplefmt[n];
//}
//
//static const AVProfile *go_av_profile_get(const AVProfile *profile, int n)
//{
//  if (profile == NULL)
//  {
//    return NULL;
//  }
//  return &profile[n];
//}
//
//static int *go_av_int_get(int *arr, int n)
//{
//  if (arr == NULL)
//  {
//    return NULL;
//  }
//  return &arr[n];
//}
//
//static uint64_t *go_av_uint64_get(uint64_t *arr, int n)
//{
//  if (arr == NULL)
//  {
//    return NULL;
//  }
//  return &arr[n];
//}
//
//static const char* get_list_at(const char **list, const int idx)
//{
//  return list[idx];
//}
//
// int GO_AVCODEC_VERSION_MAJOR = LIBAVCODEC_VERSION_MAJOR;
// int GO_AVCODEC_VERSION_MINOR = LIBAVCODEC_VERSION_MINOR;
// int GO_AVCODEC_VERSION_MICRO = LIBAVCODEC_VERSION_MICRO;
//
// #cgo pkg-config: libavcodec libavutil
import "C"
import (
	"demo/pkg/avutil"
	"errors"
	"unsafe"
)

var (
	ErrAllocationError         = errors.New("allocation error")
	ErrEncoderNotFound         = errors.New("encoder not found")
	ErrDecoderNotFound         = errors.New("decoder not found")
	ErrBitStreamFilterNotFound = errors.New("bitstreamfilter not found")
)

type CodecID C.enum_AVCodecID

const (
	CodecIDNone  CodecID = C.AV_CODEC_ID_NONE
	CodecIDMJpeg CodecID = C.AV_CODEC_ID_MJPEG
	CodecIDLJpeg CodecID = C.AV_CODEC_ID_LJPEG
)

type Capabilities int

const (
	ComplianceVeryStrict   Compliance = C.FF_COMPLIANCE_VERY_STRICT
	ComplianceStrict       Compliance = C.FF_COMPLIANCE_STRICT
	ComplianceNormal       Compliance = C.FF_COMPLIANCE_NORMAL
	ComplianceUnofficial   Compliance = C.FF_COMPLIANCE_UNOFFICIAL
	ComplianceExperimental Compliance = C.FF_COMPLIANCE_EXPERIMENTAL
)

type Compliance int

type Flags int

func Version() (int, int, int) {
	return int(C.GO_AVCODEC_VERSION_MAJOR), int(C.GO_AVCODEC_VERSION_MINOR), int(C.GO_AVCODEC_VERSION_MICRO)
}

type Packet struct {
	CAVPacket *C.AVPacket
}

func NewPacket() (*Packet, error) {
	cPkt := (*C.AVPacket)(C.av_packet_alloc())
	if cPkt == nil {
		return nil, ErrAllocationError
	}
	return NewPacketFromC(unsafe.Pointer(cPkt)), nil
}

func NewPacketFromC(cPkt unsafe.Pointer) *Packet {
	return &Packet{CAVPacket: (*C.AVPacket)(cPkt)}
}

func (pkt *Packet) Free() {
	C.av_packet_free(&pkt.CAVPacket)
}

type Codec struct {
	CAVCodec *C.AVCodec
}

func NewCodecFromC(cCodec unsafe.Pointer) *Codec {
	return &Codec{CAVCodec: (*C.AVCodec)(cCodec)}
}

func FindDecoderByID(codecID CodecID) *Codec {
	cCodec := C.avcodec_find_decoder((C.enum_AVCodecID)(codecID))
	if cCodec == nil {
		return nil
	}

	return NewCodecFromC(unsafe.Pointer(cCodec))
}

func FindEncoderByID(codecID CodecID) *Codec {
	cCodec := C.avcodec_find_encoder((C.enum_AVCodecID)(codecID))
	if cCodec == nil {
		return nil
	}
	return NewCodecFromC(unsafe.Pointer(cCodec))
}

type Context struct {
	CAVCodecContext *C.AVCodecContext
	*avutil.OptionAccessor
}

func NewContextWithCodec(codec *Codec) (*Context, error) {
	var cCodec *C.AVCodec
	if codec != nil {
		cCodec = codec.CAVCodec
	}
	cCtx := C.avcodec_alloc_context3(cCodec)
	if cCtx == nil {
		return nil, errors.New("allocation error")
	}
	return NewContextFromC(unsafe.Pointer(cCtx)), nil
}

func NewContextFromC(cCtx unsafe.Pointer) *Context {
	return &Context{
		CAVCodecContext: (*C.AVCodecContext)(cCtx),
		OptionAccessor:  avutil.NewOptionAccessor(cCtx, false),
	}
}

func (ctx *Context) Free() {
	if ctx.CAVCodecContext != nil {
		C.avcodec_free_context(&ctx.CAVCodecContext)
	}
}

func (ctx *Context) OpenWithCodec(codec *Codec, options *avutil.Dictionary) error {
	var cCodec *C.AVCodec
	if codec != nil {
		cCodec = codec.CAVCodec
	}
	var cOptions **C.AVDictionary
	if options != nil {
		cOptions = (**C.AVDictionary)(options.Pointer())
	}
	code := C.avcodec_open2(ctx.CAVCodecContext, cCodec, cOptions)
	if code < 0 {
		return avutil.NewErrorFromCode(avutil.ErrorCode(code))
	}
	return nil
}

func (ctx *Context) CodecType() avutil.MediaType {
	return (avutil.MediaType)(ctx.CAVCodecContext.codec_type)
}

func (ctx *Context) Flags() Flags {
	return Flags(ctx.CAVCodecContext.flags)
}

func (ctx *Context) Codec() *Codec {
	if ctx.CAVCodecContext.codec == nil {
		return nil
	}
	return NewCodecFromC(unsafe.Pointer(ctx.CAVCodecContext.codec))
}

func (ctx *Context) CopyTo(dst *Context) error {
	// added in lavc 57.33.100
	parameters, err := NewCodecParameters()
	if err != nil {
		return err
	}
	defer parameters.Free()
	cParams := (*C.AVCodecParameters)(unsafe.Pointer(parameters.CAVCodecParameters))
	code := C.avcodec_parameters_from_context(cParams, ctx.CAVCodecContext)
	if code < 0 {
		return avutil.NewErrorFromCode(avutil.ErrorCode(code))
	}
	code = C.avcodec_parameters_to_context(dst.CAVCodecContext, cParams)
	if code < 0 {
		return avutil.NewErrorFromCode(avutil.ErrorCode(code))
	}
	return nil
}

func (ctx *Context) CodecID() CodecID {
	return (CodecID)(ctx.CAVCodecContext.codec_id)
}

func (ctx *Context) Width() int {
	return int(ctx.CAVCodecContext.width)
}

func (ctx *Context) Height() int {
	return int(ctx.CAVCodecContext.height)
}

func (ctx *Context) PixelFormat() avutil.PixelFormat {
	return (avutil.PixelFormat)(ctx.CAVCodecContext.pix_fmt)
}

func (ctx *Context) TimeBase() *avutil.Rational {
	return avutil.NewRationalFromC(unsafe.Pointer(&ctx.CAVCodecContext.time_base))
}

func (ctx *Context) SetCodec(codec *Codec) {
	var cCodec *C.AVCodec
	if codec != nil {
		cCodec = codec.CAVCodec
	}
	ctx.CAVCodecContext.codec = cCodec
}

func (ctx *Context) SetCodecType(codecType avutil.MediaType) {
	ctx.CAVCodecContext.codec_type = (C.enum_AVMediaType)(codecType)
}

func (ctx *Context) SetFlags(flags Flags) {
	ctx.CAVCodecContext.flags = (C.int)(flags)
}

func (ctx *Context) SetWidth(width int) {
	ctx.CAVCodecContext.width = (C.int)(width)
}

func (ctx *Context) SetHeight(height int) {
	ctx.CAVCodecContext.height = (C.int)(height)
}

func (ctx *Context) SetPixelFormat(pixelFormat avutil.PixelFormat) {
	ctx.CAVCodecContext.pix_fmt = (C.enum_AVPixelFormat)(pixelFormat)
}

func (ctx *Context) SetTimeBase(timeBase *avutil.Rational) {
	ctx.CAVCodecContext.time_base.num = (C.int)(timeBase.Numerator())
	ctx.CAVCodecContext.time_base.den = (C.int)(timeBase.Denominator())
}

func (ctx *Context) SetStrictStdCompliance(compliance Compliance) {
	ctx.CAVCodecContext.strict_std_compliance = (C.int)(compliance)
}

type CodecParameters struct {
	CAVCodecParameters *C.AVCodecParameters
}

func NewCodecParameters() (*CodecParameters, error) {
	cPkt := (*C.AVCodecParameters)(C.avcodec_parameters_alloc())
	if cPkt == nil {
		return nil, errors.New("allocation error")
	}
	return NewCodecParametersFromC(unsafe.Pointer(cPkt)), nil
}

func NewCodecParametersFromC(cPSD unsafe.Pointer) *CodecParameters {
	return &CodecParameters{CAVCodecParameters: (*C.AVCodecParameters)(cPSD)}
}

func (cParams *CodecParameters) Free() {
	C.avcodec_parameters_free(&cParams.CAVCodecParameters)
}
