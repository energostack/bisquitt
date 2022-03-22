MSCS := $(wildcard img/*.msc)
MSC_PNGS := $(subst .msc,.png,$(MSCS))

all: $(MSC_PNGS)

%.png: %.msc
	mscgen -T png -o $@ $<
