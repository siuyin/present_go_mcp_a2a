import os
def str(envar: str, dflt: str) -> str:
    ret = dflt
    e = os.getenv(envar)
    if e != None:
        ret = e
    return ret
