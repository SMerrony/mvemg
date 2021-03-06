# Copyright ©2017-2020 Steve Merrony

# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.

# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
# THE SOFTWARE.

# Go build constraint used to choose between physical (hardware) and logical(sofware) emulation
# routines in some packages
TAGS = -tags physical

# Go parameters
GOCMD = go
GOBUILD = ${GOCMD}	build ${TAGS}
GOCLEAN = ${GOCMD}	clean -cache -testcache
GOTEST = ${GOCMD}	test ${TAGS} -v -race -cover # ./... github.com/SMerrony/dgemug/...
GOGET = ${GOCMD}	get
BINARY_NAME = mvemg

INSTRGEN = dginstr
INSTRSRC = ${HOME}/go/src/github.com/SMerrony/dgemug/cmd/dginstr/dginstrs.csv
INSTRGO = ../dgemug/mvcpu/instructionDefinitions.go

CLEANLOGS = rm -f logs/*.log *.pprof

# Values for program version etc.
# VERSION = 0.1
# BUILD = `git rev-parse HEAD`
# RELEASETYPE = Prerelease

# LDFLAGS = -ldflags "-X main.Version=${VERSION} -X main.Build=${BUILD} -X main.ReleaseType=${RELEASETYPE}"

all: generate build test
generate:
	${INSTRGEN} -action=makego -cputype=mv -csv=${INSTRSRC} -go=${INSTRGO}
build: 
	${GOBUILD} ${LDFLAGS} -o ${BINARY_NAME} -v
buildrace: 
	${GOBUILD} -race ${LDFLAGS} -o ${BINARY_NAME} -v	
test: 
	${GOTEST} ./...
	${GOTEST} ../dgemug/devices/...
	${GOTEST} ../dgemug/memory/...
	${GOTEST} ../dgemug/mvcpu/...
clean: 
	${GOCLEAN}
	${CLEANLOGS}
	rm -f ${BINARY_NAME} debug debug.test
cleanlogs:
	${CLEANLOGS}
run:
	${GOBUILD} -o ${BINARY_NAME} -v ./...
	./${BINARY_NAME}
deps:
	${GOGET} github.com/SMerrony/dgemug/...
	${GOGET} github.com/SMerrony/simhtape/...
