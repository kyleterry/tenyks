from setuptools import setup, find_packages
import sys, os

version = '0.1.1'

setup(name='tenyks',
      version=version,
      description="redis powered IRC bot",
      long_description="""\
""",
      classifiers=[], # Get strings from http://pypi.python.org/pypi?%3Aaction=list_classifiers
      keywords='irc bot redis',
      author='Kyle Terry',
      author_email='kyle@kyleterry.com',
      url='',
      license='GPL v2',
      packages=find_packages(exclude=['ez_setup', 'examples', 'tests']),
      include_package_data=True,
      zip_safe=False,
      install_requires=[
          'gevent',
          'redis',
          'nose',
      ],
      entry_points="""
      # -*- Entry points: -*-
      """,
      )
