

SPEECHSDK_ROOT:=speechsdk/MicrosoftCognitiveServicesSpeech.xcframework/macos-arm64_x86_64

# CHECK_FOR_SPEECHSDK := $(shell test -f $(SPEECHSDK_ROOT)/MicrosoftCognitiveServicesSpeech.framework/MicrosoftCognitiveServicesSpeech && echo Success)
# ifneq ("$(CHECK_FOR_SPEECHSDK)","Success")
#   $(error Please set SPEECHSDK_ROOT to point to your extracted Speech SDK, $$SPEECHSDK_ROOT/MicrosoftCognitiveServicesSpeech.framework/MicrosoftCognitiveServicesSpeech should exist.)
# endif

LIBS:=-framework MicrosoftCognitiveServicesSpeech
all: textToWav

# Note: to run, DYLD_FRAMEWORK_PATH should point to $SPEECHSDK_ROOT.
textToWav: textToWav.cpp
	g++ $< -o $@ \
	    --std=c++14 \
	    $(patsubst %,-F%, $(SPEECHSDK_ROOT)) \
	    $(LIBS)
