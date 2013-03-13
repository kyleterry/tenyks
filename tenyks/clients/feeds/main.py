import sqlite3
from os.path import join

import logging

import gevent.monkey
import feedparser

from tenyks.client import Client, run_client
import tenyks.config as config

gevent.monkey.patch_all()


class TenyksFeeds(Client):

    message_filters = {
        'add_feed': r'add feed (.*)',
        'list_feeds': r'list feeds',
        'del_feed': r'delete feed (.*)',
    }
    direct_only = True

    def __init__(self):
        super(TenyksFeeds, self).__init__()
        self.create_tables(self.fetch_cursor())

    def fetch_cursor(self):
        db_file = '{name}.db'.format(name=self.name)
        conn = sqlite3.connect(join(config.DATA_WORKING_DIR, db_file))
        return conn.cursor()

    def recurring(self):
        self.logger.debug('Fetching feeds')
        cur = self.fetch_cursor()
        for channel in self.get_channels(cur):
            connection = self.get_connection(cur, channel)
            for feed_obj in self.feeds_by_channel(cur, channel):
                self.feed_handler(cur, feed_obj, channel, connection)

    def feed_handler(self, cur, feed_obj, channel, connection):
        feed = feedparser.parse(feed_obj[1])
        title = feed['feed']['title']
        self.logger.debug('Looking for entries in {feed}'.format(
            feed=feed_obj[1]))
        for i, entry in enumerate(feed['entries']):
            message = '[{feed}] {title} - {link}'.format(feed=title,
                title=entry['title'],
                link=entry['link'])
            if not self.entry_exists(cur, entry['id'], feed_obj) and i < 6:
                data = {
                    'version': 1,
                    'type': 'privmsg',
                    'client': self.name,
                    'payload': message,
                    'irc_channel': channel[1],
                    'connection_name': connection[1]
                }
                self.send(message, data)
                self.create_entry(cur, entry['id'], feed_obj)

    def handle_add_feed(self, data, match):
        if data['nick_from'] in data['admins']:
            feed_url = match.groups()[0]
            self.logger.debug('add_feed: {feed}'.format(feed=feed_url))
            cur = self.fetch_cursor()
            connection = self.get_or_create_connection(cur,
                    data['connection_name'])
            channel = self.get_or_create_channel(cur,
                    connection, data['irc_channel'])
            feed = self.get_or_create_feed(cur, channel, feed_url)
            self.send('{nick_from}: {feed_url} is a go!'.format(
                        nick_from=data['nick_from'], feed_url=feed_url),
                        data=data)

    def handle_del_feed(self, data, match):
        if data['nick_from'] in data['admins']:
            feed_url = match.groups()[0]
            self.logger.debug('del_feed: {feed}'.format(feed=feed_url))
            cur = self.fetch_cursor()
            connection = self.get_or_create_connection(cur,
                data['connection_name'])
            channel = self.get_or_create_channel(cur,
                connection, data['irc_channel'])
            if self.feed_exists(cur, feed_url, channel):
                self.delete_feed(cur, feed_url, channel)


    def handle_list_feeds(self, data, match):
        self.logger.debug('list_feeds')
        cur = self.fetch_cursor()
        connection = self.get_or_create_connection(
                cur, data['connection_name'])
        channel = self.get_or_create_channel(
                cur, connection, data['irc_channel'])
        feed_sql = """
            SELECT * FROM feed
            WHERE channel_id = ?"""
        result = cur.execute(feed_sql, (channel[0],)).fetchone()
        if not result:
            self.send('{nick}: No feeds.'.format(nick=data['nick_from']), data)
        else:
            self.send('{nick}: Feeds for this channel:'.format(
                        nick=data['nick_from']), data)
            for i, feed in enumerate(cur.execute(feed_sql, (channel[0],))):
                self.send('{i}. {feed_url}'.format(i=i+1,
                    feed_url=feed[1]), data)

    def get_or_create_connection(self, cur, name):
        connection_sql = """
            SELECT *
            FROM connection
            WHERE connection_name = ?"""
        result = cur.execute(connection_sql, (name,))
        connection = result.fetchone()
        if not connection:
            result = cur.execute("""
                INSERT INTO connection (connection_name)
                VALUES (?)
            """, (name,))
            result = cur.execute(connection_sql, (name,))
            cur.connection.commit()
            connection = result.fetchone()
        return connection

    def get_or_create_channel(self, cur, connection, channel_name):
        channel_sql = """
            SELECT * FROM channel
            WHERE channel = ?
            AND connection_id = ?"""
        result = cur.execute(channel_sql, (channel_name, connection[0]))
        channel = result.fetchone()
        if not channel:
            result = cur.execute("""
            INSERT INTO channel (channel, connection_id)
            VALUES (?, ?)""", (channel_name, connection[0]))
            result = cur.execute(channel_sql, (channel_name, connection[0]))
            cur.connection.commit()
            channel = result.fetchone()
        return channel


    def get_or_create_feed(self, cur, channel, feed_url):
        feed_sql = """
            SELECT * FROM feed
            WHERE channel_id = ?
            AND feed_url = ?
        """
        result = cur.execute(feed_sql, (channel[0], feed_url))
        feed = result.fetchone()
        if not feed:
            result = cur.execute("""
            INSERT INTO feed (channel_id, feed_url)
            VALUES (?, ?)""", (channel[0], feed_url))
            cur.connection.commit()
            feed = result.fetchone()
        return feed

    def get_channels(self, cur):
        result = cur.execute("""
            SELECT * FROM channel
        """)
        return result.fetchall()

    def get_connection(self, cur, channel):
        return cur.execute("""
            SELECT * FROM connection
            WHERE id = ?""", (channel[2],)).fetchone()

    def feeds_by_channel(self, cur, channel):
        result = cur.execute("""
            SELECT * FROM feed
            WHERE channel_id = ?""", (channel[0],))
        return result.fetchall()

    def feed_exists(self, cur, feed_url, channel):
        result = cur.execute("""
            SELECT * FROM feed
            WHERE channel_id = ?
            AND feed_url = ?
        """, (channel[0], feed_url))
        return result.fetchone() is not None

    def delete_feed(self, cur, feed_url, channel):
        result = cur.execute("""
            DELETE FROM feed
            WHERE channel_id = ?
            AND feed_url = ?
        """, (channel[0], feed_url))
        cur.connection.commit()

    def entry_exists(self, cur, entry_id, feed_obj):
        result = cur.execute("""
            SELECT * FROM entry
            WHERE entry_key = ?
            AND feed_id = ?
        """, (entry_id, feed_obj[0]))
        return result.fetchone() is not None

    def create_entry(self, cur, entry_id, feed_obj):
        result = cur.execute("""
            INSERT INTO entry (entry_key, feed_id)
            VALUES (?, ?)
        """, (entry_id, feed_obj[0]))
        result.connection.commit()

    def create_tables(self, cur):
        table_sql = """
        CREATE TABLE IF NOT EXISTS connection (
            id INTEGER PRIMARY KEY,
            connection_name TEXT
        );
        CREATE TABLE IF NOT EXISTS channel (
            id INTEGER PRIMARY KEY,
            channel TEXT,
            connection_id INTEGER,
            FOREIGN KEY(connection_id)
                REFERENCES connection(id)
        );
        CREATE TABLE IF NOT EXISTS feed (
            id INTEGER PRIMARY KEY,
            feed_url TEXT,
            channel_id INTEGER,
            FOREIGN KEY(channel_id)
                REFERENCES channel(id)
        );
        CREATE TABLE IF NOT EXISTS entry (
            id INTEGER PRIMARY KEY,
            entry_key TEXT,
            feed_id INTEGER,
            FOREIGN KEY(feed_id)
                REFERENCES entry(id)
        );
        """
        cur.executescript(table_sql)

if __name__ == '__main__':
    feed = TenyksFeeds()
    run_client(feed)
