from tenyks.client import TenyksClient, run_service, CLIENT_TYPE_SERVICE


class TenyksSearch(TenyksClient):

    client_name = 'search'
    hear = r'^search (!{1}[gdw]{1}) (.*)$'

    def run(self):
        while True:
            data = self.heard_queue.get()
            print data

if __name__ == '__main__':
    search = TenyksSearch()
    run_service(search)
