import sqlite3
from os.path import join

import logging
logger = logging.getLogger()

import gevent.monkey
import feedparser

from tenyks.client import Client, run_client
import tenyks.config as config

gevent.monkey.patch_all()


class TenyksFeeds(Client):

    message_filter = r'add feed (.*)'

    def __init__(self):
        super(TenyksFeeds, self).__init__()
        self.create_tables(self.fetch_cursor())

    def fetch_cursor(self):
        db_file = '{name}.db'.format(name=self.name)
        conn = sqlite3.connect(join(config.DATA_WORKING_DIR, db_file))
        return conn.cursor()

    def recurring(self):
        logger.debug('recurring called')

    def handle(self, data, match):
        feed_url = match.groups()[0]
        logger.debug('trying to add {feed_url}'.format(feed_url=feed_url))

    def create_tables(self, cur):
        channel_sql = """
        CREATE TABLE IF NOT EXISTS channel (
            id INTEGER PRIMARY KEY,
            channel TEXT,
            connection_name TEXT
        )
        """
        feed_sql = """
        CREATE TABLE IF NOT EXISTS feed (
            id INTEGER PRIMARY KEY,
            channel_id INTEGER,
            FOREIGN KEY(channel_id)
                REFERENCES channel(id)
        )
        """
        cur.execute(channel_sql)
        cur.execute(feed_sql)

if __name__ == '__main__':
    feed = TenyksFeeds()
    run_client(feed)
