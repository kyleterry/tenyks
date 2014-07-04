# came from: http://stackoverflow.com/questions/6811902/import-arbitrary-named-file-as-a-python-module-without-generating-bytecode-file
import sys
import imp
import contextlib


@contextlib.contextmanager
def preserve_value(namespace, name):
    """ A context manager to preserve, then restore, the specified binding.

        :param namespace: The namespace object (e.g. a class or dict)
            containing the name binding.
        :param name: The name of the binding to be preserved.
        :yield: None.

        When the context manager is entered, the current value bound to `name`
        in `namespace` is saved. When the context manager is exited, the
        binding is re-established to the saved value.

        """
    saved_value = getattr(namespace, name)
    yield
    setattr(namespace, name, saved_value)


def make_module_from_file(module_name, module_filepath):
    """ Make a new module object from the source code in specified file.

        :param module_name: The name of the resulting module object.
        :param module_filepath: The filesystem path to open for
            reading the module's Python source.
        :return: The module object.

        The Python import mechanism is not used. No cached bytecode
        file is created, and no entry is placed in `sys.modules`.

        """
    py_source_open_mode = 'U'
    py_source_description = (b".py", py_source_open_mode, imp.PY_SOURCE)

    with open(module_filepath, py_source_open_mode) as module_file:
        with preserve_value(sys, 'dont_write_bytecode'):
            sys.dont_write_bytecode = True
            module = imp.load_module(
                module_name,
                module_file,
                module_filepath,
                py_source_description)

    return module
