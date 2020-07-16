---
title: "Ubidots Setup"
description: ""
weight: 1
---

Create a function by following these instructions to prepare Ubidots setup for the integration with {{% tts %}}. This function will handle the conversion of incoming messages from JSON format to the one that is compatible with Ubidots.

<!--more-->

Log in to your Ubidots account and find the **Devices** tab in the upper part of your dashboard. In its drop-down list, choose **Functions**.

When redirected to the **UbiFunctions** page, create a new function with the **Create Function** button.

On the left, give a **Name** to your function, select the **POST** method and choose **Python 3.6** for **Runtime**.

In this example, we will use the following Python function, which receives a JSON payload from {{% tts %}}, extracts the `decoded_payload`, and posts received data to Ubidots.

>Note: to find your **Token**, click on your avatar in the upper right corner and select **API Credentials**. Be sure to use your **Token** and not your **API Key**.

```
import requests
import json
import time

BASE_URL = "https://industrial.api.ubidots.com"
TOKEN = "····" # Enter a token here

def main(args):
    # Prining args from TTI
    print(f'[INFO] Args from TTI:\n {args}')

    # Parsing data
    payload = parse_tti_data(args)
    dev_label = tti_dev_eui(args)
    print(f'[INFO] Parsed data:\n {payload}')
    print(f'[INFO] TTI Dev_EUI data:\n {dev_label}')

    # Posting to Ubidots
    req = update_device(dev_label, payload, TOKEN)
    print(f'[INFO] Request to Ubidots Status code: {req.status_code}')
    print(f'[INFO] Request ti Ubidots JSON:\n {req.json()}')

    return {
        'status_code': req.status_code,
        'response_json': req.json()
    }


def parse_tti_data(data):
    return data['uplink_message']['decoded_payload']


def tti_dev_eui(data):
    return data['end_device_ids']['device_id']


def update_device(device, payload, token):
    """
    Updates device with payload
    """
    url = "{}/api/v1.6/devices/{}".format(BASE_URL, device)
    headers = {"X-Auth-Token": token, "Content-Type": "application/json"}
    req = create_request(url, headers, attempts=5, request_type="post", data=payload)
    return req


def create_request(url, headers, attempts, request_type, data=None):
    """
    Function to make a request to the server
    """
    request_func = getattr(requests, request_type)
    kwargs = {"url": url, "headers": headers}
    if request_type == "post" or request_type == "patch":
        kwargs["json"] = data
    try:
        req = request_func(**kwargs)
        status_code = req.status_code
        time.sleep(1)
        while status_code >= 400 and attempts < 5:
            req = request_func(**kwargs)
            status_code = req.status_code
            attempts += 1
            time.sleep(1)
        return req
    except Exception as e:
        print("[ERROR] There was an error with the request, details:")
        print(e)
        return None
```

After modifying the function code with your token, click the **Make it live** button. 

Your function will be assigned an **HTTPS Endpoint URL**. Copy this URL in order to use it later as a part of setup on {{% tts %}}. 

{{< figure src="creating-function.png" alt="Creating a UbiFunction" >}}

Your function is now ready to process the incoming messages from {{% tts %}}.