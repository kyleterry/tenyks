from tenyks.client import TenyksClient, run_service, CLIENT_TYPE_SERVICE


class TenyksBrain(TenyksClient):

    service_name = 'brain'
    client_type = CLIENT_TYPE_SERVICE

    def run(self):
        for data in self.input_queue.get():
            print data


if __name__ == '__main__':
    brain = TenyksBrain()
    run_service(brain)
