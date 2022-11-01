package main

import (
	"demo/pkg/avcodec"
	"demo/pkg/avfilter"
	"demo/pkg/avformat"
	"demo/pkg/avutil"
	"log"
)

type context struct {
	// decoding
	decFmt    *avformat.Context
	decStream *avformat.Stream
	decCodec  *avcodec.Context
	decPkt    *avcodec.Packet
	decFrame  *avutil.Frame
	srcFilter *avfilter.Context

	// encoding
	encFmt     *avformat.Context
	encStream  *avformat.Stream
	encCodec   *avcodec.Context
	encIO      *avformat.IOContext
	encPkt     *avcodec.Packet
	encFrame   *avutil.Frame
	sinkFilter *avfilter.Context

	// transcoding
	filterGraph *avfilter.Graph
}

func main() {
	ctx, err := newContext()
	if err != nil {
		log.Fatalf("Failed to create context: %v\n", err)
	}

	defer ctx.free()
	openInput(ctx)
	openOutput(ctx)
}

func newContext() (*context, error) {
	ctx := &context{}
	if err := ctx.alloc(); err != nil {
		ctx.free()
		return nil, err
	}
	return ctx, nil
}

func openInput(ctx *context) {
	var err error

	// open format (container) context
	ctx.decFmt, err = avformat.NewContextForInput()
	if err != nil {
		log.Fatalf("Failed to open input context: %v\n", err)
	}

	// set some options for opening file
	options := avutil.NewDictionary()
	defer options.Free()
	if err := options.Set("scan_all_pmts", "1"); err != nil {
		log.Fatalf("Failed to set input options: %v\n", err)
	}

	// open file for decoding
	inputFileName := "./file/2.mp4"
	if err := ctx.decFmt.OpenInput(inputFileName, nil, options); err != nil {
		log.Fatalf("Failed to open input file: %v\n", err)
	}

	// initialize context with stream information
	if err := ctx.decFmt.FindStreamInfo(nil); err != nil {
		log.Fatalf("Failed to find stream info: %v\n", err)
	}

	// dump streams to standard output
	ctx.decFmt.Dump(0, inputFileName, false)

	// prepare first video stream for decoding
	openFirstInputVideoStream(ctx)
}

func openFirstInputVideoStream(ctx *context) {
	var err error

	// find first video stream
	if ctx.decStream = firstVideoStream(ctx.decFmt); ctx.decStream == nil {
		log.Fatalf("Could not find a video stream. Aborting...\n")
	}

	codecCtx := ctx.decStream.CodecContext()
	codec := avcodec.FindDecoderByID(codecCtx.CodecID())
	if codec == nil {
		log.Fatalf("Could not find decoder: %v\n", codecCtx.CodecID())
	}
	if ctx.decCodec, err = avcodec.NewContextWithCodec(codec); err != nil {
		log.Fatalf("Failed to create codec context: %v\n", err)
	}
	if err := codecCtx.CopyTo(ctx.decCodec); err != nil {
		log.Fatalf("Failed to copy codec context: %v\n", err)
	}
	if err := ctx.decCodec.SetInt64Option("refcounted_frames", 1); err != nil {
		log.Fatalf("Failed to copy codec context: %v\n", err)
	}
	if err := ctx.decCodec.OpenWithCodec(codec, nil); err != nil {
		log.Fatalf("Failed to open codec: %v\n", err)
	}

	// we need a video filter to push the decoded frames to
	ctx.srcFilter = addFilter(ctx, "buffer", "in")

	//给上边的 过滤器上下文 srcFilter设置参数
	if err = ctx.srcFilter.SetImageSizeOption("video_size", ctx.decCodec.Width(), ctx.decCodec.Height()); err != nil {
		log.Fatalf("Failed to set filter option: %v\n", err)
	}
	if err = ctx.srcFilter.SetPixelFormatOption("pix_fmt", ctx.decCodec.PixelFormat()); err != nil {
		log.Fatalf("Failed to set filter option: %v\n", err)
	}
	if err = ctx.srcFilter.SetRationalOption("time_base", ctx.decCodec.TimeBase()); err != nil {
		log.Fatalf("Failed to set filter option: %v\n", err)
	}
	log.Println("video_size:", ctx.decCodec.Width(), ctx.decCodec.Height(),
		"pix_fmt:", ctx.decCodec.PixelFormat(),
		"time_base:", ctx.decCodec.TimeBase())

	//根据上边的设置来的过滤器上下文来初始化过滤器
	if err = ctx.srcFilter.Init(); err != nil {
		log.Fatalf("Failed to initialize buffer filter: %v\n", err)
	}
}

func firstVideoStream(fmtCtx *avformat.Context) *avformat.Stream {
	for _, stream := range fmtCtx.Streams() {
		switch stream.CodecContext().CodecType() {
		case avutil.MediaTypeVideo:
			return stream
		}
	}
	return nil
}

func openOutput(ctx *context) {
	var err error
	outputFileName := "./file/out2.mp4"
	// guess format given output filename
	fmt := avformat.GuessOutputFromFileName(outputFileName)
	if fmt == nil {
		log.Fatalf("Failed to guess output for output file: %s\n", outputFileName)
	}
	if ctx.encFmt, err = avformat.NewContextForOutput(fmt); err != nil {
		log.Fatalf("Failed to open output context: %v\n", err)
	}

	// prepare first video stream for encoding
	openOutputVideoStream(ctx, fmt)

	if fmt.Flags()&avformat.FlagNoFile != 0 {
		return
	}

	// prepare I/O
	flags := avformat.IOFlagWrite
	if ctx.encIO, err = avformat.OpenIOContext(outputFileName, flags, nil, nil); err != nil {
		log.Fatalf("Failed to open I/O context: %v\n", err)
	}
	ctx.encFmt.SetIOContext(ctx.encIO)
}

func openOutputVideoStream(ctx *context, fmt *avformat.Output) {
	var err error
	outputFileName := "./file/out2.mp4"
	ctx.encStream, err = ctx.encFmt.NewStreamWithCodec(nil)
	if err != nil {
		log.Fatalf("Failed to open output video stream: %v\n", err)
	}
	codecCtx := ctx.encStream.CodecContext()
	codecCtx.SetCodecType(avutil.MediaTypeVideo)
	codecID := fmt.GuessCodecID(outputFileName, codecCtx.CodecType())
	codec := avcodec.FindEncoderByID(codecID)
	if codec == nil {
		log.Fatalf("Could not find encoder: %v\n", codecID)
	}
	if ctx.encCodec, err = avcodec.NewContextWithCodec(codec); err != nil {
		log.Fatalf("Failed to create codec context: %v\n", err)
	}
	ctx.encCodec.SetCodecType(codecCtx.CodecType())

	// we need a video filter to pull the encoded frames from
	ctx.sinkFilter = addFilter(ctx, "buffersink", "out")
	if err = ctx.sinkFilter.Init(); err != nil {
		log.Fatalf("Failed to initialize buffersink filter: %v\n", err)
	}
	if err = ctx.srcFilter.Link(0, ctx.sinkFilter, 0); err != nil {
		log.Fatalf("Failed to link filters: %v\n", err)
	}
	if err = ctx.filterGraph.Config(); err != nil {
		log.Fatalf("Failed to config filter graph: %v\n", err)
	}

	sinkPads := ctx.sinkFilter.Inputs()
	sinkPad := sinkPads[0]
	ctx.encCodec.SetWidth(sinkPad.Width())
	ctx.encCodec.SetHeight(sinkPad.Height())
	ctx.encCodec.SetPixelFormat(sinkPad.PixelFormat())
	ctx.encCodec.SetTimeBase(ctx.decCodec.TimeBase())
	ctx.encCodec.SetStrictStdCompliance(avcodec.ComplianceNormal)

	if fmt.Flags()&avformat.FlagGlobalHeader != 0 {
		// ctx.encCodec.SetFlags(ctx.encCodec.Flags() | avcodec.FlagGlobalHeader)
		ctx.encCodec.SetFlags(ctx.encCodec.Flags())
	}

	if err = ctx.encCodec.OpenWithCodec(codec, nil); err != nil {
		log.Fatalf("Failed to open codec: %v\n", err)
	}
	if err := ctx.encCodec.CopyTo(ctx.encStream.CodecContext()); err != nil {
		log.Fatalf("Failed to copy codec context: %v\n", err)
	}
	ctx.encStream.SetTimeBase(ctx.encCodec.TimeBase())
	ctx.encStream.CodecContext().SetCodec(ctx.encCodec.Codec())
}

func addFilter(ctx *context, name, id string) *avfilter.Context {
	filter := avfilter.FindFilterByName(name)
	if filter == nil {
		log.Fatalf("Could not find %s/%s filter\n", name, id)
	}
	fctx, err := ctx.filterGraph.AddFilter(filter, id)
	if err != nil {
		log.Fatalf("Failed to add %s/%s filter: %v\n", name, id, err)
	}
	return fctx
}

func (ctx *context) alloc() error {
	var err error
	if ctx.decPkt, err = avcodec.NewPacket(); err != nil {
		return err
	}
	if ctx.decFrame, err = avutil.NewFrame(); err != nil {
		return err
	}
	if ctx.encPkt, err = avcodec.NewPacket(); err != nil {
		return err
	}
	if ctx.encFrame, err = avutil.NewFrame(); err != nil {
		return err
	}
	if ctx.filterGraph, err = avfilter.NewGraph(); err != nil {
		return err
	}
	return nil
}

func (ctx *context) free() {
	if ctx.encIO != nil {
		ctx.encIO.Close()
		ctx.encIO = nil
	}
	if ctx.encFmt != nil {
		ctx.encFmt.Free()
		ctx.encFmt = nil
	}
	if ctx.filterGraph != nil {
		ctx.filterGraph.Free()
		ctx.filterGraph = nil
	}
	if ctx.encPkt != nil {
		ctx.encPkt.Free()
		ctx.encPkt = nil
	}
	if ctx.encFrame != nil {
		ctx.encFrame.Free()
		ctx.encFrame = nil
	}
	if ctx.decPkt != nil {
		ctx.decPkt.Free()
		ctx.decPkt = nil
	}
	if ctx.decFrame != nil {
		ctx.decFrame.Free()
		ctx.decFrame = nil
	}
	if ctx.decCodec != nil {
		ctx.decCodec.Free()
		ctx.decCodec = nil
	}
	if ctx.decFmt != nil {
		ctx.decFmt.CloseInput()
		ctx.decFmt.Free()
		ctx.decFmt = nil
	}
}
