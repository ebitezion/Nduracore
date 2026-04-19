#!/usr/bin/env bash

mockgen -source=interfaces/random.go -destination=testutil/random_mock.go -package=interfaces