FROM bakingbad/michelson-kernel

COPY build/sandbox/jupyter/notebooks /home/jupyter/notebooks

ENTRYPOINT [ "jupyter",  "notebook", "--ip=0.0.0.0", "--no-browser", "--no-mathjax"]

CMD ["--NotebookApp.allow_origin=*", "--NotebookApp.port=8008", "--NotebookApp.default_url=/tree/notebooks", "--NotebookApp.disable_check_xsrf=True", "--NotebookApp.token="]
