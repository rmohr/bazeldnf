def rpms():
    rpm(
        name = "a-0__1.2.3.myarch",
        sha256 = "1234",
        urls = [
            "a/something/a",
            "b/something/a",
            "c/something/a",
        ],
    )
    rpm(
        name = "a-0__2.3.4.myarch",
        sha256 = "1234",
        urls = [
            "e/something/a",
            "f/something/a",
            "g/something/a",
        ],
    )
    rpm(
        name = "b-0__2.3.4.myarch",
        sha256 = "1234",
        urls = [
            "a/something/b",
            "b/something/b",
            "c/something/b",
        ],
    )

    rpm(
        name = "test.rpm",
        urls = ["http://something.rpm"],
    )
