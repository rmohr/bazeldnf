"provider for unit-test"

def rpms():
    rpm(
        name = "test.rpm",
        urls = ["http://something.rpm"],
    )
