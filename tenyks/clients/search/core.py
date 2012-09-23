from tenyks.client import TenyksClient, run_service, CLIENT_TYPE_SERVICE


class TenyksSearch(TenyksClient):

    client_name = 'brain'

    def run(self):
        print 'running'
        for match in self.hear(r'^search (!{1}[gdw]{1}) (.*)$'):
            search_type, search_query = match.groups()
            print search_type, search_query


if __name__ == '__main__':
    search = TenyksSearch()
    run_service(search)
