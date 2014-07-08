#!/usr/bin/env python
import os
import sys

from ledis import __version__

try:
    from setuptools import setup
    from setuptools.command.test import test as TestCommand

    class PyTest(TestCommand):
        def finalize_options(self):
            TestCommand.finalize_options(self)
            self.test_args = []
            self.test_suite = True

        def run_tests(self):
            # import here, because outside the eggs aren't loaded
            import pytest
            errno = pytest.main(self.test_args)
            sys.exit(errno)

except ImportError:

    from distutils.core import setup
    PyTest = lambda x: x


setup(
    name='ledis',
    version=__version__,
    description='Python client for ledisdb key-value database',
    long_description='Python client for ledisdb key-value database',
    url='https://github.com/siddontang/ledisdb',
    keywords=['ledis', 'key-value store'],
    license='MIT',
    packages=['ledis'],
    tests_require=['pytest>=2.5.0'],
    cmdclass={'test': PyTest},
    classifiers=[
        'Development Status :: 5 - Production/Stable',
        'Environment :: Console',
        'Intended Audience :: Developers',
        'License :: OSI Approved :: MIT License',
        'Operating System :: OS Independent',
        'Programming Language :: Python',
        'Programming Language :: Python :: 2.6',
        'Programming Language :: Python :: 2.7',
    ]
)
