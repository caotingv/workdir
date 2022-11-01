package avformat

//#include <libavutil/avutil.h>
//#include <libavutil/avstring.h>
//#include <libavcodec/avcodec.h>
//#include <libavformat/avformat.h>
//
//#ifdef AVFMT_FLAG_FAST_SEEK
//#define GO_AVFMT_FLAG_FAST_SEEK AVFMT_FLAG_FAST_SEEK
//#else
//#define GO_AVFMT_FLAG_FAST_SEEK 0
//#endif
//
//static const AVStream *go_av_streams_get(const AVStream **streams, unsigned int n)
//{
//  return streams[n];
//}
//
//static AVDictionary **go_av_alloc_dicts(int length)
//{
//  size_t size = sizeof(AVDictionary*) * length;
//  return (AVDictionary**)av_malloc(size);
//}
//
//static void go_av_dicts_set(AVDictionary** arr, unsigned int n, AVDictionary *val)
//{
//  arr[n] = val;
//}
//
//
// int GO_AVFORMAT_VERSION_MAJOR = LIBAVFORMAT_VERSION_MAJOR;
// int GO_AVFORMAT_VERSION_MINOR = LIBAVFORMAT_VERSION_MINOR;
// int GO_AVFORMAT_VERSION_MICRO = LIBAVFORMAT_VERSION_MICRO;
//
//typedef int (*AVFormatContextIOOpenCallback)(struct AVFormatContext *s, AVIOContext **pb, const char *url, int flags, AVDictionary **options);
//typedef void (*AVFormatContextIOCloseCallback)(struct AVFormatContext *s, AVIOContext *pb);
//
// #cgo pkg-config: libavformat libavutil
import "C"
import (
	"errors"
	"unsafe"

	"demo/pkg/avcodec"
	"demo/pkg/avutil"
)

var (
	ErrAllocationError     = errors.New("allocation error")
	ErrInvalidArgumentSize = errors.New("invalid argument size")
)

type Flags int

const (
	FlagNoFile       Flags = C.AVFMT_NOFILE
	FlagGlobalHeader Flags = C.AVFMT_GLOBALHEADER
)

type IOFlags int

const (
	IOFlagRead      IOFlags = C.AVIO_FLAG_READ
	IOFlagWrite     IOFlags = C.AVIO_FLAG_WRITE
	IOFlagReadWrite IOFlags = C.AVIO_FLAG_READ_WRITE
	IOFlagNonblock  IOFlags = C.AVIO_FLAG_NONBLOCK
	IOFlagDirect    IOFlags = C.AVIO_FLAG_DIRECT
)

func Version() (int, int, int) {
	return int(C.GO_AVFORMAT_VERSION_MAJOR), int(C.GO_AVFORMAT_VERSION_MINOR), int(C.GO_AVFORMAT_VERSION_MICRO)
}

type Input struct {
	CAVInputFormat *C.AVInputFormat
}

func FindInputByShortName(shortName string) *Input {
	cShortName := C.CString(shortName)
	defer C.free(unsafe.Pointer(cShortName))
	cInput := C.av_find_input_format(cShortName)
	if cInput == nil {
		return nil
	}
	return NewInputFromC(unsafe.Pointer(cInput))
}

func NewInputFromC(cInput unsafe.Pointer) *Input {
	return &Input{CAVInputFormat: (*C.AVInputFormat)(cInput)}
}

type Output struct {
	CAVOutputFormat *C.AVOutputFormat
}

func NewOutputFromC(cOutput unsafe.Pointer) *Output {
	return &Output{CAVOutputFormat: (*C.AVOutputFormat)(cOutput)}
}

func GuessOutputFromFileName(fileName string) *Output {
	cFileName := C.CString(fileName)
	defer C.free(unsafe.Pointer(cFileName))
	cOutput := C.av_guess_format(nil, cFileName, nil)
	if cOutput == nil {
		return nil
	}
	return NewOutputFromC(unsafe.Pointer(cOutput))
}

func (f *Output) GuessCodecID(filename string, mediaType avutil.MediaType) avcodec.CodecID {
	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))
	return (avcodec.CodecID)(C.av_guess_codec(f.CAVOutputFormat, nil, cFilename, nil, C.enum_AVMediaType(mediaType)))
}

type Stream struct {
	CAVStream *C.AVStream
}

func NewStreamFromC(cStream unsafe.Pointer) *Stream {
	return &Stream{CAVStream: (*C.AVStream)(cStream)}
}

func (s *Stream) CodecContext() *avcodec.Context {
	if s.CAVStream.codec == nil {
		return nil
	}
	return avcodec.NewContextFromC(unsafe.Pointer(s.CAVStream.codec))
}

func (s *Stream) SetTimeBase(timeBase *avutil.Rational) {
	s.CAVStream.time_base.num = (C.int)(timeBase.Numerator())
	s.CAVStream.time_base.den = (C.int)(timeBase.Denominator())
}

type Context struct {
	CAVFormatContext *C.AVFormatContext
}

func NewContextForInput() (*Context, error) {
	cCtx := C.avformat_alloc_context()
	if cCtx == nil {
		return nil, errors.New("allocation error")
	}
	return NewContextFromC(unsafe.Pointer(cCtx)), nil
}

func NewContextForOutput(output *Output) (*Context, error) {
	var cCtx *C.AVFormatContext
	code := C.avformat_alloc_output_context2(&cCtx, output.CAVOutputFormat, nil, nil)
	if code < 0 {
		return nil, avutil.NewErrorFromCode(avutil.ErrorCode(code))
	}
	return NewContextFromC(unsafe.Pointer(cCtx)), nil
}

func NewContextFromC(cCtx unsafe.Pointer) *Context {
	return &Context{CAVFormatContext: (*C.AVFormatContext)(cCtx)}
}

func (ctx *Context) SetIOContext(ioCtx *IOContext) {
	var cIOCtx *C.AVIOContext
	if ioCtx != nil {
		cIOCtx = ioCtx.CAVIOContext
	}
	ctx.CAVFormatContext.pb = cIOCtx
}

func (ctx *Context) OpenInput(fileName string, input *Input, options *avutil.Dictionary) error {
	cFileName := C.CString(fileName)
	defer C.free(unsafe.Pointer(cFileName))
	var cInput *C.AVInputFormat
	if input != nil {
		cInput = input.CAVInputFormat
	}
	var cOptions **C.AVDictionary
	if options != nil {
		cOptions = (**C.AVDictionary)(options.Pointer())
	}
	code := C.avformat_open_input(&ctx.CAVFormatContext, cFileName, cInput, cOptions)
	if code < 0 {
		return errors.New("avformat open input failed")
	}
	return nil
}

func (ctx *Context) CloseInput() {
	C.avformat_close_input(&ctx.CAVFormatContext)
}

func (ctx *Context) Free() {
	if ctx.CAVFormatContext != nil {
		defer C.avformat_free_context(ctx.CAVFormatContext)
		ctx.CAVFormatContext = nil
	}
}

func (ctx *Context) FindStreamInfo(options []*avutil.Dictionary) error {
	code := C.avformat_find_stream_info(ctx.CAVFormatContext, nil)
	if code < 0 {
		return errors.New("not find stream info")
	}
	return nil
}

func (ctx *Context) NumberOfStreams() uint {
	return uint(ctx.CAVFormatContext.nb_streams)
}

func (ctx *Context) Dump(streamIndex int, url string, isOutput bool) {
	var cIsOutput C.int
	if isOutput {
		cIsOutput = C.int(1)
	}
	cURL := C.CString(url)
	defer C.free(unsafe.Pointer(cURL))
	C.av_dump_format(ctx.CAVFormatContext, C.int(streamIndex), cURL, cIsOutput)
}

func (ctx *Context) Streams() []*Stream {
	count := ctx.NumberOfStreams()
	if count <= 0 {
		return nil
	}
	streams := make([]*Stream, 0, count)
	for i := uint(0); i < count; i++ {
		cStream := C.go_av_streams_get(ctx.CAVFormatContext.streams, C.uint(i))
		stream := NewStreamFromC(unsafe.Pointer(cStream))
		streams = append(streams, stream)
	}
	return streams
}

func (ctx *Context) NewStreamWithCodec(codec *avcodec.Codec) (*Stream, error) {
	var cCodec *C.AVCodec
	if codec != nil {
		cCodec = (*C.AVCodec)(unsafe.Pointer(codec.CAVCodec))
	}
	cStream := C.avformat_new_stream(ctx.CAVFormatContext, cCodec)
	if cStream == nil {
		return nil, errors.New("allocation error")
	}
	return NewStreamFromC(unsafe.Pointer(cStream)), nil
}

// func (ctx *Context) SetFileName(fileName string) {
// 	cFileName := C.CString(fileName)
// 	defer C.free(unsafe.Pointer(cFileName))
// 	C.av_strlcpy(&ctx.CAVFormatContext.filename[0], cFileName, C.sizeOfAVFormatContextFilename)
// }

type IOContext struct {
	CAVIOContext *C.AVIOContext
}

func OpenIOContext(url string, flags IOFlags, cb *IOInterruptCallback, options *avutil.Dictionary) (*IOContext, error) {
	cURL := C.CString(url)
	defer C.free(unsafe.Pointer(cURL))
	var cCb *C.AVIOInterruptCB
	if cb != nil {
		cCb = cb.CAVIOInterruptCB
	}
	var cOptions **C.AVDictionary
	if options != nil {
		cOptions = (**C.AVDictionary)(options.Pointer())
	}
	var cCtx *C.AVIOContext
	code := C.avio_open2(&cCtx, cURL, (C.int)(flags), cCb, cOptions)
	if code < 0 {
		return nil, avutil.NewErrorFromCode(avutil.ErrorCode(code))
	}
	return NewIOContextFromC(unsafe.Pointer(cCtx)), nil

}

func NewIOContextFromC(cCtx unsafe.Pointer) *IOContext {
	return &IOContext{CAVIOContext: (*C.AVIOContext)(cCtx)}
}

func (f *Output) Flags() Flags {
	return Flags(f.CAVOutputFormat.flags)
}

func (ctx *IOContext) Close() error {
	if ctx.CAVIOContext != nil {
		code := C.avio_closep(&ctx.CAVIOContext)
		if code < 0 {
			return avutil.NewErrorFromCode(avutil.ErrorCode(code))
		}
	}
	return nil
}

type IOInterruptCallback struct {
	CAVIOInterruptCB *C.AVIOInterruptCB
}

func NewIOInterruptCallbackFromC(cb unsafe.Pointer) *IOInterruptCallback {
	return &IOInterruptCallback{CAVIOInterruptCB: (*C.AVIOInterruptCB)(cb)}
}
