#!/usr/bin/env python
from setuptools import setup, find_packages
import sys, os

version = '0.2.1'

packages = [
    'tenyks',
    'tenyks.client',
]

setup(name='tenyks',
      version=version,
      description="Redis powered IRC bot",
      long_description="""\
""",
      classifiers=[], # Get strings from http://pypi.python.org/pypi?%3Aaction=list_classifiers
      keywords='irc bot redis',
      author='Kyle Terry',
      author_email='kyle@kyleterry.com',
      url='https://github.com/kyleterry/tenyks',
      license='LICENSE',
      packages=packages,
      package_dir={'tenyks': 'tenyks'},
      package_data={'tenyks': ['*.pem', '*.dist', 'client/*.dist']},
      include_package_data=True,
      zip_safe=False,
      test_suite='tests',
      install_requires=[
          'gevent',
          'redis',
          'nose',
          'unittest2',
          'clint',
          'peewee',
          'jinja2',
      ],
      entry_points={
          'console_scripts': [
              'tenyks = tenyks.core:main',
              'tenyks-mkconfig = tenyks.config:make_config',
              'tenyks-client-mkconfig = tenyks.client.config:make_config'
          ]
      },
      )
