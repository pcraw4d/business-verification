"""
Setup script for KYB Platform Risk Assessment Service Python SDK
"""

from setuptools import setup, find_packages

with open("README.md", "r", encoding="utf-8") as fh:
    long_description = fh.read()

with open("requirements.txt", "r", encoding="utf-8") as fh:
    requirements = [line.strip() for line in fh if line.strip() and not line.startswith("#")]

setup(
    name="kyb-sdk",
    version="1.0.0",
    author="KYB Platform",
    author_email="support@kyb-platform.com",
    description="KYB Platform Risk Assessment Service Python SDK",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://github.com/kyb-platform/python-sdk",
    project_urls={
        "Bug Reports": "https://github.com/kyb-platform/python-sdk/issues",
        "Source": "https://github.com/kyb-platform/python-sdk",
        "Documentation": "https://docs.kyb-platform.com",
    },
    packages=find_packages(),
    classifiers=[
        "Development Status :: 5 - Production/Stable",
        "Intended Audience :: Developers",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Programming Language :: Python :: 3.12",
        "Topic :: Software Development :: Libraries :: Python Modules",
        "Topic :: Office/Business :: Financial",
        "Topic :: Security",
    ],
    python_requires=">=3.8",
    install_requires=requirements,
    extras_require={
        "dev": [
            "pytest>=7.0.0",
            "pytest-cov>=4.0.0",
            "black>=22.0.0",
            "flake8>=5.0.0",
            "mypy>=1.0.0",
        ],
        "docs": [
            "sphinx>=5.0.0",
            "sphinx-rtd-theme>=1.0.0",
        ],
    },
    keywords="kyb risk-assessment compliance business-verification api sdk",
    include_package_data=True,
    zip_safe=False,
)
