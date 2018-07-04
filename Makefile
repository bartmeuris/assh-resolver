PLATFORMS := linux/amd64 windows/amd64/.exe darwin/amd64
APP := assh-resolver

temp = $(subst /, ,$@)
tos   = $(word 1, $(temp))
tarch = $(word 2, $(temp))
ext = $(word 3, $(temp))

all: $(PLATFORMS)

$(PLATFORMS):
	CGO_ENABLED=0 GOOS=$(tos) GOARCH=$(tarch) go build -o '$(APP)_$(tos)-$(tarch)$(ext)' .

.PHONY: release $(PLATFORMS)
