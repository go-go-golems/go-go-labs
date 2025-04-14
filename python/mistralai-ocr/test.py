import os
from mistralai import Mistral

api_key = os.environ["MISTRAL_API_KEY"]
client = Mistral(api_key=api_key)

ocr_response = client.ocr.process(
    model="mistral-ocr-latest",
    document={
        "type": "document_url",
        # "document_url": "https://arxiv.org/pdf/2201.04234",
        "document_url": "https://dspace.mit.edu/bitstream/handle/1721.1/6913/AITR-474.pdf",
    },
    include_image_base64=True,
)

import json

with open("aitr_response.json", "w") as f:
    json.dump(ocr_response.model_dump(), f)
