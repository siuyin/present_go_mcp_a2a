from langchain_ollama import ChatOllama
import os,sys

sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__),"..")))
from dflt import env

def main():
    model = env.str("MODEL","gemma3:1b")
    system = env.str("SYS","Provide concise responses.")
    prompt = env.str("PROMPT", "What is the meaning of life?")
    host = env.str("OLLAMA_HOST","http://localhost:11434")
    print(f"MODEL={model} HOST={host} SYS={system} PROMPT={prompt}")

    think = env.str("THINK","aloud")
    tv:any = False
    if (think == "aloud"):
        tv = None
    elif (think == "true"):
        tv = True
    else:
        tv = False

    llm = ChatOllama(
            model = model,
            validate_model_on_init = True,
            temperature = 0.1,
            reasoning=tv,
            )

    messages = [
            ("system", system),
            ("user", prompt),
            ]

    try:
        os.remove("/tmp/j")
    except:
        pass
    f = open("/tmp/j","a")
    for chunk in llm.stream(messages):
        if "reasoning_content" in chunk.additional_kwargs:
            print(chunk.additional_kwargs["reasoning_content"],end="", file=f)
        else:
            print(chunk.text(),end="")

    print()
    f.close()

if (__name__) == "__main__":
    main()
