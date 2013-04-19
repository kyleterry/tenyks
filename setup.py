#!/usr/bin/env python
from setuptools import setup, find_packages
import sys, os

version = '0.1.24'

packages = [
    'tenyks',
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
      package_data={'tenyks': ['*.pem']},
      include_package_data=True,
      zip_safe=False,
      install_requires=[
          'gevent',
          'redis',
          'nose',
          'unittest2',
          'clint',
      ],
      entry_points={
          'console_scripts': [
              'tenyks = tenyks.core:main',
              'tenyksmkconfig = tenyks.config:make_config'
          ]
      },
      )
