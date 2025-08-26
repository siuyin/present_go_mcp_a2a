import os
import sys
from langchain_ollama import ChatOllama
from langchain_core.messages import BaseMessage,ChatMessage
from typing import Any,Sequence

sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__),"..")))
from dflt import env

def main() -> None:
    model:str = env.string("MODEL","gemma3:1b")
    system:str = env.string("SYS","Provide concise responses.")
    prompt:str = env.string("PROMPT", "What is the meaning of life?")
    host:str = env.string("OLLAMA_HOST","http://localhost:11434")
    print(f"MODEL={model} HOST={host} SYS={system} PROMPT={prompt}")

    # See: https://python.langchain.com/api_reference/ollama/chat_models/langchain_ollama.chat_models.ChatOllama.html#chatollama
    messages_from_langchain_example = [
            ("system", system),
            ("user", prompt),
            ]
    messages:Sequence[BaseMessage] = [
            ChatMessage(role="system",content=system),
            ChatMessage(role="user",content=prompt),
            ]

    chat = MyChat(host,model,messages) # change messages arg
    chat.complete()

class MyChat():
    def __init__(self, host:str, model:str, messages:Sequence[BaseMessage]) -> None:
        self.host:str =host
        self.model:str =model
        self.msgs:Sequence[BaseMessage] = messages

        think:str = env.string("THINK","aloud")
        tv:Any = False
        if (think == "aloud"):
            tv = None
        elif (think == "true"):
            tv = True
        else:
            tv = False
        self.cl:ChatOllama = ChatOllama(model=model,validate_model_on_init=True,temperature=0.1,reasoning=tv)

    def complete(self) -> None:
        try:
            os.remove("/tmp/j")
        except:
            pass
        f = open("/tmp/j","a")
        for chunk in self.cl.stream(self.msgs):
            if "reasoning_content" in chunk.additional_kwargs:
                print(chunk.additional_kwargs["reasoning_content"],end="", file=f)
            else:
                print(chunk.text(),end="")

        print()
        f.close()

def newMyChat(host:str, model:str, messages:Sequence[BaseMessage]) -> MyChat:
    return MyChat(host,model,messages)

if (__name__) == "__main__":
    main()
