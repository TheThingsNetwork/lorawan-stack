---
title: "Azure Setup"
description: ""
weight: 1
---

Follow the instructions in this section to create a function in Azure and prepare the setup for integration with {{% tts %}}.

<!--more-->

>Note: you can find more detailed instructions for creating Azure functions in the [official Azure documentation](https://docs.microsoft.com/en-us/azure/azure-functions/functions-create-first-azure-function).

First, navigate to the Azure services dashboard and click on **Function App** service.

{{< figure src="azure-services-dashboard.png" alt="Function app in Azure services dashboard" >}}

Once you are in **Function App** creation wizard, all the options that you will need for this guide are in the **Basics** tab. 

Fill in the **Subscription** and **Resource Group** fields according to your preferences. If it is your first time using the Azure, you will have to create a new resource group.

Next, you have to provide a globally unique **Function App name**. After deploying your function, you can check if it is running by navigating to `https://{function_app_name}.azurewebsites.net/`.

Choose **Code** for **Publish** to publish your code files (the alternative would be publishing a Docker container). 

Since you can use different programming languages for your function, you have to choose an appropriate **Runtime stack**. In this guide, we use a **C#** function, therefore we choose **.NET Core** runtime stack. Also, you need to choose the **Version** of the installed runtime. The most recent version will automatically be suggested to you.

Choose a **Region** nearest to you, or near other services that your functions access. 

{{< figure src="creating-a-function-app.png" alt="Creating Azure function app" >}}

When finished, click on **Review + create** button, and then once again on **Create** button. At this point, you should see something like this:

{{< figure src="successfully-deployed-function-app.png" alt="Function app deployment complete" >}}

To view and manage your function app, select **Go to resource**.

In the next step, you will be creating a function inside your function app. From the left hand **Functions** menu, choose **Functions** and click on **Add** button. 

In the **New Function** pop-up menu, select **HTTP trigger**. Give your function a recognizable name and choose **Function** for **Authorization level**. 

>Note: **Authorization level** defines whether the function requires an API key and if so, what kind (function or master key). Depending on this, an API key may be a part of the webhook URL. 

Select **Create Function** to finish. 

{{< figure src="http-trigger-function-creation.png" alt="Creating HTTP trigger function" >}}

After creating a function, click on it and select **Code + Test** in the **Developer** menu on the left. Modify the function code as shown below in order to show incoming messages in JSON format. 

Also make sure to copy the **function URL**, because you will need it when creating a Webhook integration on {{% tts %}}. 

{{< figure src="http-trigger-function-code.png" alt="HTTP trigger function code" >}}

Finally, expand the **Logs** in the bottom and click on the **Start** button.
