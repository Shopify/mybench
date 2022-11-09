#!/usr/bin/env python3

from setuptools import setup, find_packages

setup(
  name="mybench_analysis",
  version="0.1",
  description="Support package for analyzing mybench data",
  author="Shuhao Wu",
  url="https://github.com/Shopify/mybench",
  packages=list(find_packages()),
)
