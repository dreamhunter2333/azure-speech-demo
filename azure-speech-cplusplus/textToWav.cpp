#include <iostream>
#include <fstream>
#include <string>
#include <MicrosoftCognitiveServicesSpeech/speechapi_cxx.h>

using namespace std;
using namespace Microsoft::CognitiveServices::Speech;

void synthesizeSpeech(
    const SPXSTRING &endpoint, const SPXSTRING &subscription, const SPXSTRING &voiceName,
    const SPXSTRING &text, const SPXSTRING &wavFile)
{
    auto config = SpeechConfig::FromEndpoint(endpoint, subscription);
    config->SetSpeechSynthesisVoiceName(voiceName);
    auto synthesizer = SpeechSynthesizer::FromConfig(config, NULL);
    auto result = synthesizer->SpeakTextAsync(text).get();
    auto stream = AudioDataStream::FromResult(result);
    stream->SaveToWavFile(wavFile);
}
