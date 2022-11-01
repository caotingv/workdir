package avfilter

//#include <libavutil/avutil.h>
//#include <libavutil/opt.h>
//#include <libavutil/error.h>
//#include <libavfilter/avfilter.h>
//#include <libavfilter/buffersrc.h>
//#include <libavfilter/buffersink.h>
//
//#ifdef AV_BUFFERSRC_FLAG_NO_COPY
//#define GO_AV_BUFFERSRC_FLAG_NO_COPY AV_BUFFERSRC_FLAG_NO_COPY
//#else
//#define GO_AV_BUFFERSRC_FLAG_NO_COPY 0
//#endif
//
//static const AVFilterLink *go_av_links_get(const AVFilterLink **links, unsigned int n)
//{
//  return links[n];
//}
//
//static const int GO_AVERROR(int e)
//{
//  return AVERROR(e);
//}
//
// int GO_AVFILTER_VERSION_MAJOR = LIBAVFILTER_VERSION_MAJOR;
// int GO_AVFILTER_VERSION_MINOR = LIBAVFILTER_VERSION_MINOR;
// int GO_AVFILTER_VERSION_MICRO = LIBAVFILTER_VERSION_MICRO;
//
//#define GO_AVFILTER_AUTO_CONVERT_ALL ((unsigned)AVFILTER_AUTO_CONVERT_ALL)
//#define GO_AVFILTER_AUTO_CONVERT_NONE ((unsigned)AVFILTER_AUTO_CONVERT_NONE)
//
//#cgo pkg-config: libavfilter libavutil
import "C"

import (
	"errors"
	"fmt"
	"unsafe"

	"demo/pkg/avutil"
)

var (
	ErrAllocationError = errors.New("allocation error")
)

type Flags int

const (
	FlagDynamicInputs           Flags = C.AVFILTER_FLAG_DYNAMIC_INPUTS
	FlagDynamicOutputs          Flags = C.AVFILTER_FLAG_DYNAMIC_OUTPUTS
	FlagSliceThreads            Flags = C.AVFILTER_FLAG_SLICE_THREADS
	FlagSupportTimelineGeneric  Flags = C.AVFILTER_FLAG_SUPPORT_TIMELINE_GENERIC
	FlagSupportTimelineInternal Flags = C.AVFILTER_FLAG_SUPPORT_TIMELINE_INTERNAL
	FlagSupportTimeline         Flags = C.AVFILTER_FLAG_SUPPORT_TIMELINE
)

type BufferSrcFlags C.int

const (
	BufferSrcFlagNoCheckFormat BufferSrcFlags = C.AV_BUFFERSRC_FLAG_NO_CHECK_FORMAT
	BufferSrcFlagNoCopy        BufferSrcFlags = C.GO_AV_BUFFERSRC_FLAG_NO_COPY
	BufferSrcFlagPush          BufferSrcFlags = C.AV_BUFFERSRC_FLAG_PUSH
	BufferSrcFlagKeepRef       BufferSrcFlags = C.AV_BUFFERSRC_FLAG_KEEP_REF
)

type GraphAutoConvertFlags uint

const (
	GraphAutoConvertFlagAll  GraphAutoConvertFlags = C.GO_AVFILTER_AUTO_CONVERT_ALL
	GraphAutoConvertFlagNone GraphAutoConvertFlags = C.GO_AVFILTER_AUTO_CONVERT_NONE
)

func Version() (int, int, int) {
	return int(C.GO_AVFILTER_VERSION_MAJOR), int(C.GO_AVFILTER_VERSION_MINOR), int(C.GO_AVFILTER_VERSION_MICRO)
}

type Filter struct {
	CAVFilter *C.AVFilter
}

func NewFilterFromC(cFilter unsafe.Pointer) *Filter {
	return &Filter{CAVFilter: (*C.AVFilter)(cFilter)}
}

// 通过名称初始化过滤器. 看起来应该是用来填充数据用的
func FindFilterByName(name string) *Filter {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	cFilter := C.avfilter_get_by_name(cName)
	if cFilter == nil {
		return nil
	}
	return NewFilterFromC(unsafe.Pointer(cFilter))
}

type Link struct {
	CAVFilterLink *C.AVFilterLink
}

func NewLinkFromC(cLink unsafe.Pointer) *Link {
	return &Link{CAVFilterLink: (*C.AVFilterLink)(cLink)}
}

func (l *Link) Width() int {
	return int(l.CAVFilterLink.w)
}

func (l *Link) Height() int {
	return int(l.CAVFilterLink.h)
}

func (l *Link) PixelFormat() avutil.PixelFormat {
	return avutil.PixelFormat(l.CAVFilterLink.format)
}

type Context struct {
	CAVFilterContext *C.AVFilterContext
	*avutil.OptionAccessor
}

func NewContextFromC(cCtx unsafe.Pointer) *Context {
	return &Context{
		CAVFilterContext: (*C.AVFilterContext)(cCtx),
		OptionAccessor:   avutil.NewOptionAccessor(cCtx, false),
	}
}

func (ctx *Context) Init() error {
	options := avutil.NewDictionary()
	defer options.Free()
	return ctx.InitWithDictionary(options)
}

/* Now initialize the filter; we pass NULL options, since we have already
 * set all the options above. */
func (ctx *Context) InitWithDictionary(options *avutil.Dictionary) error {
	var cOptions **C.AVDictionary
	if options != nil {
		cOptions = (**C.AVDictionary)(options.Pointer())
	}
	fmt.Println(&cOptions)
	code := C.avfilter_init_dict(ctx.CAVFilterContext, cOptions)
	if code < 0 {
		return avutil.NewErrorFromCode(avutil.ErrorCode(code))
	}
	return nil
}

func (ctx *Context) Link(srcPad uint, dst *Context, dstPad uint) error {
	cSrc := ctx.CAVFilterContext
	cDst := dst.CAVFilterContext
	code := C.avfilter_link(cSrc, C.uint(srcPad), cDst, C.uint(dstPad))
	if code < 0 {
		return avutil.NewErrorFromCode(avutil.ErrorCode(code))
	}
	return nil
}

func (ctx *Context) Inputs() []*Link {
	count := ctx.NumberOfInputs()
	if count <= 0 {
		return nil
	}
	links := make([]*Link, 0, count)
	for i := uint(0); i < count; i++ {
		cLink := C.go_av_links_get(ctx.CAVFilterContext.inputs, C.uint(i))
		link := NewLinkFromC(unsafe.Pointer(cLink))
		links = append(links, link)
	}
	return links
}

func (ctx *Context) NumberOfInputs() uint {
	return uint(ctx.CAVFilterContext.nb_inputs)
}

type Graph struct {
	CAVFilterGraph *C.AVFilterGraph
}

//分配过滤器图形控件
func NewGraph() (*Graph, error) {
	cGraph := C.avfilter_graph_alloc()
	if cGraph == nil {
		return nil, ErrAllocationError
	}
	return NewGraphFromC(unsafe.Pointer(cGraph)), nil
}

func NewGraphFromC(cGraph unsafe.Pointer) *Graph {
	return &Graph{CAVFilterGraph: (*C.AVFilterGraph)(cGraph)}
}

//通过已有的 CAVFilterGraph 过滤器图形生成,名称是 cName,返回过滤器上下文
func (g *Graph) AddFilter(filter *Filter, name string) (*Context, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	cCtx := C.avfilter_graph_alloc_filter(g.CAVFilterGraph, filter.CAVFilter, cName)
	if cCtx == nil {
		return nil, ErrAllocationError
	}
	return NewContextFromC(unsafe.Pointer(cCtx)), nil
}

func (g *Graph) Free() {
	if g.CAVFilterGraph != nil {
		C.avfilter_graph_free(&g.CAVFilterGraph)
	}
}

func (g *Graph) Config() error {
	code := C.avfilter_graph_config(g.CAVFilterGraph, nil)
	if code < 0 {
		return avutil.NewErrorFromCode(avutil.ErrorCode(code))
	}
	return nil
}
