.. _getting-started:

===============
Getting started
===============

This guide walks you through installing and running an example benchmark
implemented with mybench.

To download and compile ``examplebench``:

.. code-block:: shell-session

   $ git clone https://github.com/Shopify/mybench.git
   $ cd mybench
   $ make examplebench

This will compile ``examplebench`` into a folder called ``build``. To run
``examplebench``, you must first seed the database:

.. code-block:: shell-session

   $ build/examplebench \
      --host=mysql.host \
      --user=username \
      --pass=password \
      --load

You need to replace the ``mysql.host`` with the host or IP address of MySQL,
``username`` with the username you can connect with, and ``password`` with the
password you can connect to. This will load the database with 1 million rows of
data (in the table ``example_table`` in the database ``mybench``).

Once this is done, you can then run the benchmark:


.. code-block:: shell-session

   $ build/examplebench \
      --host=mysql.host \
      --user=username \
      --pass=password \
      --bench \
      --eventrate=10000

The default event rate for examplebench is 1000 event/s split evenly between its
various workloads. The ``--eventrate=10000`` option overrides this, specifying
an event rate 10x the default, automatically distributed among the defined workloads.
You can then go to https://localhost:8005 to see the real-time monitoring UI. 
This should show something similar to:

.. image:: images/screenshot.png

By default, the test will go on indefinitely (a fixed duration can be specified
using a config option). Pressing ``CTRL+C`` will abort the test. The data will
be saved into a file called ``data.sqlite``.

TODO: a bit more about how to use the post processing scripts.
