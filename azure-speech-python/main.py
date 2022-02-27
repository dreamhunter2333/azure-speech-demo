import click
import azure.cognitiveservices.speech as speechsdk


@click.command()
@click.option('--endpoint', '-e', 'endpoint', required=True, type=str)
@click.option('--subscription', '-s', 'subscription', required=True, type=str)
@click.option('--voicename', '-v', 'voice_name', default="zh-CN-XiaomoNeural", type=str)
@click.option('--text', '-t', 'text', required=True, type=str)
@click.option('--out', '-o', 'out', required=False, type=str)
def main(endpoint, subscription, voice_name, text, out):
    click.echo("endpoint {}, subscription {}, voice_name {}, text {}, out {}".format(
        endpoint, subscription, voice_name, text, out
    ))
    speech_config = speechsdk.SpeechConfig(
        endpoint=endpoint,
        subscription=subscription
    )
    speech_config.speech_synthesis_voice_name = voice_name
    speech_synthesizer = speechsdk.SpeechSynthesizer(
        speech_config=speech_config, audio_config=None
    ) if out else speechsdk.SpeechSynthesizer(
        speech_config=speech_config
    )
    result = speech_synthesizer.speak_text_async(text).get()
    if not out:
        click.echo("output to the default speaker")
        return
    stream = speechsdk.AudioDataStream(result)
    stream.save_to_wav_file(out)
    click.echo("save_to_wav_file: {}".format(out))


main()
