import mpd
import gevent.monkey
gevent.monkey.patch_all()

from tenyks.client import Client, run_client


class TenyksMpdMusic(Client):

    message_filters = {
        'play': r'play music',
        'pause': r'pause music',
        'next': r'next song',
        'random_toggle': r'toggle random',
        'currentsong': r'current song',
        'stats': r'music stats',
    }
    direct_only = True

    def get_client(self):
        client = mpd.MPDClient()
        client.timeout = 10
        client.idletimeout = None
        client.connect('localhost', 6600)
        return client

    def handle_play(self, data, match):
        client = self.get_client()
        client.play()
        client.disconnect()

    def handle_pause(self, data, match):
        client = self.get_client()
        client.pause()
        client.disconnect()

    def handle_next(self, data, match):
        client = self.get_client()
        client.next()
        client.disconnect()

    def handle_random_toggle(self, data, match):
        client = self.get_client()
        status = client.status()
        if status['random'] == '0':
            client.random(1)
        else:
            client.random(0)
        client.disconnect()

    def handle_currentsong(self, data, match):
        client = self.get_client()
        status = client.currentsong()
        if 'albumartist' in status:
            artist = status['albumartist']
        else:
            artist = status['artist']
        message = '{nick}: {artist} - {song} ({album})'.format(
            nick=data['nick_from'], artist=artist,
            song=status['title'], album=status['album'])
        self.send(message, data)
        client.disconnect()

    def handle_stats(self, data, match):
        client = self.get_client()
        stats = client.stats()
        message = 'artists: {artists}, songs: {songs}, albums: {alb}'.format(
                artists=stats['artists'], songs=stats['songs'],
                alb=stats['albums'])
        message = '{nick}: {message}'.format(nick=data['nick_from'],
                message=message)
        self.send(message, data)
        client.disconnect()


if __name__ == '__main__':
    mpd_music = TenyksMpdMusic()
    run_client(mpd_music)
