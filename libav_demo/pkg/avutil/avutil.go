package avutil

/*
#include <libavutil/avutil.h>
#include <libavutil/channel_layout.h>
#include <libavutil/dict.h>
#include <libavutil/pixdesc.h>
#include <libavutil/opt.h>
#include <libavutil/frame.h>
#include <libavutil/parseutils.h>
#include <libavutil/common.h>
#include <libavutil/eval.h>

#ifdef AV_LOG_TRACE
#define GO_AV_LOG_TRACE AV_LOG_TRACE
#else
#define GO_AV_LOG_TRACE AV_LOG_DEBUG
#endif

#ifdef AV_PIX_FMT_XVMC_MPEG2_IDCT
#define GO_AV_PIX_FMT_XVMC_MPEG2_IDCT AV_PIX_FMT_XVMC_MPEG2_MC
#else
#define GO_AV_PIX_FMT_XVMC_MPEG2_IDCT 0
#endif

#ifdef AV_PIX_FMT_XVMC_MPEG2_MC
#define GO_AV_PIX_FMT_XVMC_MPEG2_MC AV_PIX_FMT_XVMC_MPEG2_MC
#else
#define GO_AV_PIX_FMT_XVMC_MPEG2_MC 0
#endif

static const AVDictionaryEntry *go_av_dict_next(const AVDictionary *m, const AVDictionaryEntry *prev)
{
 return av_dict_get(m, "", prev, AV_DICT_IGNORE_SUFFIX);
}

static const int go_av_dict_has(const AVDictionary *m, const char *key, int flags)
{
 if (av_dict_get(m, key, NULL, flags) != NULL)
 {
   return 1;
 }
 return 0;
}

static int go_av_expr_parse2(AVExpr **expr, const char *s, const char * const *const_names, int log_offset, void *log_ctx)
{
 return av_expr_parse(expr, s, const_names, NULL, NULL, NULL, NULL, log_offset, log_ctx);
}

static const int go_av_errno_to_error(int e)
{
 return AVERROR(e);
}

#cgo pkg-config: libavutil
*/
import "C"
import (
	"errors"
	"unsafe"
)

var (
	ErrAllocationError     = errors.New("allocation error")
	ErrInvalidArgumentSize = errors.New("invalid argument size")
)

type MediaType C.enum_AVMediaType

const (
	MediaTypeUnknown    MediaType = C.AVMEDIA_TYPE_UNKNOWN
	MediaTypeVideo      MediaType = C.AVMEDIA_TYPE_VIDEO
	MediaTypeAudio      MediaType = C.AVMEDIA_TYPE_AUDIO
	MediaTypeData       MediaType = C.AVMEDIA_TYPE_DATA
	MediaTypeSubtitle   MediaType = C.AVMEDIA_TYPE_SUBTITLE
	MediaTypeAttachment MediaType = C.AVMEDIA_TYPE_ATTACHMENT
)

type OptionSearchFlags int

const (
	OptionSearchChildren OptionSearchFlags = C.AV_OPT_SEARCH_CHILDREN
	OptionSearchFakeObj  OptionSearchFlags = C.AV_OPT_SEARCH_FAKE_OBJ
)

func Version() (int, int, int) {
	return int(C.LIBAVUTIL_VERSION_MAJOR), int(C.LIBAVUTIL_VERSION_MINOR), int(C.LIBAVUTIL_VERSION_MICRO)
}

type Rational struct {
	CAVRational C.AVRational
}

func NewRational(numerator, denominator int) *Rational {
	r := &Rational{}
	r.CAVRational.num = C.int(numerator)
	r.CAVRational.den = C.int(denominator)
	return r
}

func NewRationalFromC(cRational unsafe.Pointer) *Rational {
	rational := (*C.AVRational)(cRational)
	return NewRational(int(rational.num), int(rational.den))
}

func (r *Rational) Numerator() int {
	return int(r.CAVRational.num)
}

func (r *Rational) Denominator() int {
	return int(r.CAVRational.den)
}

type ErrorCode int

type Error struct {
	code ErrorCode
	err  error
}

func strError(code C.int) string {
	size := C.size_t(256)
	buf := (*C.char)(C.av_mallocz(size))
	defer C.av_free(unsafe.Pointer(buf))
	if C.av_strerror(code, buf, size-1) == 0 {
		return C.GoString(buf)
	}
	return "Unknown error"
}

func NewErrorFromCode(code ErrorCode) *Error {
	return &Error{
		code: code,
		err:  errors.New(strError(C.int(code))),
	}
}

func (e *Error) Code() ErrorCode {
	return e.code
}

func (e *Error) Error() string {
	return e.err.Error()
}

type Dictionary struct {
	CAVDictionary  **C.AVDictionary
	pCAVDictionary *C.AVDictionary
}

func NewDictionary() *Dictionary {
	return NewDictionaryFromC(nil)
}

func NewDictionaryFromC(cDictionary unsafe.Pointer) *Dictionary {
	return &Dictionary{CAVDictionary: (**C.AVDictionary)(cDictionary)}
}

func (dict *Dictionary) Free() {
	C.av_dict_free(dict.pointer())
}

func (dict *Dictionary) set(key, value string, flags C.int) error {
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))
	code := C.av_dict_set(dict.pointer(), cKey, cValue, flags)
	if code < 0 {
		return NewErrorFromCode(ErrorCode(code))
	}
	return nil
}

func (dict *Dictionary) Set(key, value string) error {
	return dict.set(key, value, C.AV_DICT_MATCH_CASE)
}

func (dict *Dictionary) Pointer() unsafe.Pointer {
	return unsafe.Pointer(dict.pointer())
}

func (dict *Dictionary) pointer() **C.AVDictionary {
	if dict.CAVDictionary != nil {
		return dict.CAVDictionary
	}
	return &dict.pCAVDictionary
}

type Frame struct {
	CAVFrame *C.AVFrame
}

func NewFrame() (*Frame, error) {
	cFrame := C.av_frame_alloc()
	if cFrame == nil {
		return nil, ErrAllocationError
	}
	return NewFrameFromC(unsafe.Pointer(cFrame)), nil
}

func NewFrameFromC(cFrame unsafe.Pointer) *Frame {
	return &Frame{CAVFrame: (*C.AVFrame)(cFrame)}
}

func (f *Frame) Free() {
	if f.CAVFrame != nil {
		C.av_frame_free(&f.CAVFrame)
	}
}

type OptionAccessor struct {
	obj  unsafe.Pointer
	fake bool
}

func NewOptionAccessor(obj unsafe.Pointer, fake bool) *OptionAccessor {
	return &OptionAccessor{obj: obj, fake: fake}
}

func (oa *OptionAccessor) SetInt64Option(name string, value int64) error {
	return oa.SetInt64OptionWithFlags(name, value, OptionSearchChildren)
}

func (oa *OptionAccessor) SetInt64OptionWithFlags(name string, value int64, flags OptionSearchFlags) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	searchFlags := oa.searchFlags(flags)
	code := C.av_opt_set_int(oa.obj, cName, (C.int64_t)(value), searchFlags)
	if code < 0 {
		return NewErrorFromCode(ErrorCode(code))
	}
	return nil
}

func (oa *OptionAccessor) searchFlags(flags OptionSearchFlags) C.int {
	flags &^= OptionSearchFakeObj

	if oa.fake {
		flags |= OptionSearchFakeObj
	}
	return C.int(flags)
}

func (oa *OptionAccessor) SetImageSizeOption(name string, width, height int) error {
	return oa.SetImageSizeOptionWithFlags(name, width, height, OptionSearchChildren)
}

func (oa *OptionAccessor) SetImageSizeOptionWithFlags(name string, width, height int, flags OptionSearchFlags) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	searchFlags := oa.searchFlags(flags)
	code := C.av_opt_set_image_size(oa.obj, cName, (C.int)(width), (C.int)(height), searchFlags)
	if code < 0 {
		return NewErrorFromCode(ErrorCode(code))
	}
	return nil
}

type PixelFormat C.enum_AVPixelFormat

func (oa *OptionAccessor) SetPixelFormatOption(name string, value PixelFormat) error {
	return oa.SetPixelFormatOptionWithFlags(name, value, OptionSearchChildren)
}

func (oa *OptionAccessor) SetPixelFormatOptionWithFlags(name string, value PixelFormat, flags OptionSearchFlags) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	searchFlags := oa.searchFlags(flags)
	code := C.av_opt_set_pixel_fmt(oa.obj, cName, (C.enum_AVPixelFormat)(value), searchFlags)
	if code < 0 {
		return NewErrorFromCode(ErrorCode(code))
	}
	return nil
}

func (oa *OptionAccessor) SetRationalOption(name string, value *Rational) error {
	return oa.SetRationalOptionWithFlags(name, value, OptionSearchChildren)
}

func (oa *OptionAccessor) SetRationalOptionWithFlags(name string, value *Rational, flags OptionSearchFlags) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	searchFlags := oa.searchFlags(flags)
	code := C.av_opt_set_q(oa.obj, cName, value.CAVRational, searchFlags)
	if code < 0 {
		return NewErrorFromCode(ErrorCode(code))
	}
	return nil
}
