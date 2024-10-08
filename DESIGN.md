# The Things Stack for LoRaWAN Design Guidelines

## Introduction

We at The Things Network aim to give you an excellent experience when using our products. To make sure that our product lives up to our own expectations of an outstanding product, it is important to define a set of principles and resulting recommendations that guide and inform our design and development during all stages of production.

## Our Mission

We at The Things Network are strong believers in the power of enablement. With our great community, we have built a world-wide, decentralized, user-enabled data network – free and open for everyone. We strongly believe that the  Internet of Things is a great facilitator in solving many of the problems of our modern age, which is why wanted to empower people from around the world to make use of this potential for their own means: helping their communities, solving big problems, and simply realizing new business opportunities.

Read more about the mission and vision of The Things Network in our [manifesto](https://github.com/TheThingsNetwork/Manifesto/blob/master/Mission.md).

## Our Audience

In order to create a successful product, it is crucial to understand the types of users that are to engage with it. At The Things Network, we want everyone to be able to use our product and we regard it as part of our success, that we have been able to achieve this goal. When designing new solutions however, this constitution of our audience can pose challenges, since we don't want to oversimplify and thus turn off knowledgeable users, but we also don't want to overwhelm people with only moderate technical background.

Through a community study conducted in 2018 with 434 participants, we have a good starting point in understanding our audience. It can be divided in three distinct groups:

### Hobbyists and Community-Enthusiasts

LoRaWAN and IoT have a great attractactiveness for technology enthusiastic hobbyists and people who share our community vision from around the world. Since we at The Things Network together with our world-wide communities have brought this technology to the masses, it is natural that this user group is making up a big part of our audience. The level of LoRaWAN proficiency can vary considerably in this group.

##### Audience profile

- Varying technical proficiency, though can be assumed medium in the mean
- Value ease-of-use, assistance and accessibility of technology
- Have generally simpler use cases and demands

### Professionals and Product Developers

This audience group has a medium to high knowledge about LoRaWAN and is using The Things Network to develop PoCs or mature products. The demand for product stability and feature-completeness of our product is particularly high for this group.

##### Audience Profile

- Generally higher IoT/LoRaWAN proficiency
- Value stability, trust and security of our products
- Have elevated technical requirements

### Researchers and Educators

"LoRaWAN is here to stay" is one of the slogans often heard in our community. As the maturity of the technology progresses, researchers and educators from around the world are turning to LoRaWAN, making it part of their curriculas or investigating potentials for their companies. This audience cares a lot about accessibility and usability of our products to be able to grasp potentials and promote it further along their peers.

##### Audience Profile

- Varying technical proficiencey, though can be assumed higher in the mean
- Values stability and accessibility of technology
- Have presumably simpler use cases and demands

## Design Values

Based on the insights about our user group as well as driven by our mission, we can describe three simple design values that we use to inform all of the design decisions we make now and in the future:

### Facilitate and Enable

IoT and LoRaWAN are intricate fields with a growing amount of technicalities, driven by the aim to power of the standard further and further. We at The Things Network want to make sure that the power of LoRaWAN stays accessible for everyone and is not obstructed by the complexity behind the technology. It is important to us that people from all levels of proficiency with IoT will have the possibility to participate and profit from this amazing technology. Whether it is to simply learn and experiment – or to build a PoC for a business case, you should have all the tools at hand to get you there!

### Nurture the Power of LoRaWAN

Since we started working on the Things Stack for LoRaWAN, we wanted to harness the advancing power of LoRaWAN. The LoRa Alliance has worked hard to progress the standard further and further, making it more flexible, versatile and secure. At The Things Network, we want to make sure that these improvements will be ready at your disposal, to improve your existing use cases – or to enable new ones.

### Open and Decentralized

The Things Network owes a lot of its success to you – the community. We believe in decentralization and the spirit of open source. What we build is for everyone and everyone should feel invited to support our efforts. The Things Stack for LoRaWAN is 100% Open Source and we encourage our users to join the development and assist with whatever skill they can bring to the table. It is for this reason that we want to offer solid development, design and usage documentation, so you can have an understanding of how we work and what principles inform our decisions. This guideline is one effort in this direction.

## Design Foundation

The Things Stack for LoRaWAN employs a set of design tokens that ensure a streamlined appearance throughout the product. These values are to be used exclusively, while deviations can be valid in special cases.

Please refer to our [design library TTUI](https://thethingsindustries.github.io/design) for more information. Keep in mind though that this offers a more general design system for our entire communication, with the Console only having certain intersections. We're working on an updated design guide for the Console.

## Design Patterns

### Content and Communication

#### Voice and Tone

The console should assist with textual info that is concise, sober and formal, helping the user completing his or her task.

##### General Considerations

- Do not use casual tone in all text relating to actions or tasks of the user
- Do not use direct address of the user except for dialogs (like modals: `Are you sure you want to…`)
- Do not use end-of-sentence periods, unless the text consists of multiple sentences
- Seek to find the least amount of words without sacrificing clarity or grammar
- Use inline documentation as long as it can be short and precise (e.g. field descriptions and notifications)
- Do not write lengthy descriptions and rather refer to our documentation
- Only first-level component names of The Things Stack for LoRaWAN (e.g. `Identity Server`, `Applicaiton Server`, `Console`) are capitalized, otherwise normal sentence case is used
- Do not shorten `end device` to just `device`

##### Acknowledgement

Acknowledgement text increases the users awareness and confidence after an action has been performed successfully. For example after updating an entity, we show a toast notification confirming the success of the action. Keep in mind that by default, toast notifications dismiss themselves after a fixed amount of time (currently 4 seconds), so the text it displays should be as short as possible.

<img src="https://user-images.githubusercontent.com/5710611/79826969-0fd4cb80-83d8-11ea-9023-2ab90637624d.png" width="300px"/>

Acknowledgement text must

- be written in simple past tense
- use short and unambiguous language
- have a clear context of which action it was referring to, by mentioning the action and displaying it in temporal and/or spatial proximity of the action

| Do                 | Don't                         | Why?                  |
|--------------------|-------------------------------|-----------------------|
| End device updated | End device has been updated   | Use simple past tense |
| Gateway deleted    | Operation successful          | Avoid ambiguity       |
| Location updated   | Location updated successfully | Use short language    |

##### Error Messages

Error messages must 

- be written in simple past tense
- provide users with a clear description of the problem, if applicable
- in best case, hint towards a resolution of the problem
- have a clear context of which action it was referring to, by mentioning the action and displaying it in temporal and/or spatial proximity of the action
- provide clarity as to what consequences the error had for the desired action
- not use any word shortening using apostrophes
- if possible, not be shown in toasts

| Do                                                         | Don't                                                           | Why?                                         |
|------------------------------------------------------------|-----------------------------------------------------------------|----------------------------------------------|
| An error occurred and the end device could not be created  | There has been an error and the end device could not be created | Use simple past tense                        |
| An error occurred and the application could not be deleted | There was an error                                              | State consequences of the failure            |
| An error occurred and the changes could not be saved       | An error occurred while saving the changes                      | State consequences unambiguously             |
| Could not create Gateway: the ID is already taken          | Could not create gateway                                        | Provide reasons for the failure, if possible |
| An error occured and the organization could not be created | An error occured and the organization couldn't be created       | No word shortening                           |

##### Form Validation Messages

Form validation messages must

- be written in present tense
- avoid negative vocabulary e.g. (invalid, error, forbidden)
- provide info on what is a valid input
- be as specific as possible
- reference the field name if applicable
- use `must` and not `should`

| Do                                                            | Don't                          | Why?                                     |
|---------------------------------------------------------------|--------------------------------|------------------------------------------|
| Password must match                                           | Passwords should match         | The condition **must** be true           |
| End device ID is required                                     | Field is required              | Name the subject                         |
| Name must be less than 1024 characters                        | Name is too long               | Inform about the precise limitation      |
| Latitude must be a whole or decimal number between -90 and 90 | Latitude must be a valid value | Be as specific as possible               |


##### Action Text

Action text is text that is directly describing a related action of the User Interface. This usually refers to captions of common UI elements such as buttons, radio-buttons or checkboxes.

Action text must

- use short, imperative tense, e.g. `Save changes`
- always refer the concrete action and preferably the concrete subject
- not use title case

| Do            | Don't                           | Why?                    |
|---------------|---------------------------------|-------------------------|
| Create PubSub | Submit                          | Refer to precise action |
| Start process | Click here to start the process | Be concise              |
| Save changes  | Save Changes                    | Use sentence case       |

### Forms

By the very nature of our product, we employ a variety of forms as means to data manipulation. Good form design is a critical differentiatior for the perception of the user experience. We employ a set of rules to ensure the best compromise between maintainability and UX:

- Auto focus the first input when:
  - The form is a create form
  - It can otherwise be safely assumed that the user will fill the form in chronological order
- Use placeholders to
  - Display example values (e.g. `my-new-webhook`), when such would improve the understanding of a desired value structure
  - To inform about reasons for disabled fields
- Use field descriptions
  - To inform briefly about the purpose of the field
  - Should always be set, unless the field is *very* trivial and/or not LoRaWAN-related (such as login fields)
- Use [sensible defaults](#sensible-defaults) *wherever possible*
- Show submit errors at the top of the form and scroll the error into view if necessary
  - Do not use toasts to display form errors
- In update forms, show success messages as toasts
- In create forms, forward the user either to the newly created record or to the list of records after successful creation; do not show a toast additionally
- Use headings to separate sections and structurize the form
- [Hide form fields that are not relevant](#progressive-disclosure)
- Consider splitting up complex forms into wizards or separate, independent forms within accordions. See also [Divide and Structurize](#divide-and-structurize)

## Design Principles

Having defined our design values on a high level, we can continue to derive concrete recommendations and imperatives for our work on The Things Stack for LoRaWAN.

### Web User Interfaces

#### Handle the Power/Complexity Trade-Off

From the constitution of our audience, it becomes apparent that the technical proficiency of our users is very heterogenous. In everyday work, this makes us face a known quandary of human cumputer interaction called power/complexity trade-off. Simply forwarding all the complexity of LoRaWAN to the user will overwhelm the less knowledgeable, while simply limiting complex functionality will put power users off. To handle this problem, we aim to apply a couple of measures:

##### Progressive Disclosure

Progressive disclosure aims to only show advanced settings if the user has (explicitly or implicitly) indicated using them, thus reducing the congnitive load on the beginner user, while retaining the flexibility for those who need it.

![Progressive Disclosure](https://i.imgur.com/c71dwO3.gif)

##### Sensible Defaults

Instead of explicitly forcing the user to make an input for every possible value, we need to look into defaults that would account for 80% or more of the cases that the value is set. This would help a lot of users that just want to get started without worrying about specifics (for the moment). While the decision about such default values can be delicate, the potential with regards to the UX can be huge.
![Sensible Defaults](https://user-images.githubusercontent.com/5710611/77146162-c427bd00-6acd-11ea-8ed6-8d262bf8ef23.png) 

##### Settings Templates

While LoRaWAN can handle a countless number of use cases, there will be common scenarios that account for the majority of uses. It is important to investigate these common scenarios to be able to provide setup templates or *canned solutions* that can be applied quickly and take away the demand for extensive configuration from the user. Concretely, this could e.g. mean to provide device templates that the user can choose to prefill fields in the device creation form.

#### Divide and Structurize

![Divide and Structurize](https://user-images.githubusercontent.com/5710611/77146533-98590700-6ace-11ea-85f2-8391260c39f1.png)
Instead of overwhelming our users with complex pages that just forward the underlying technical complexity, we need to try to make the experience as digestible as possible. In practice, that means to chunk up complex processes into graspable parts:

- Split up complex processes using the wizard pattern
- Chunk up lengthy forms into collapsible accordions (see also [Progressive Disclosure](#Progressive-Disclosure))
- Use sidebars and tabs to organize different features
- Make use of whitespace as a means to avoid visual density by grouping visual elements meaningfully
- Use and maintain a clear layout grid
- Use concise inline documentation as much as possible, reference documentation where applicable
- Improve situational awareness using concise notifications, warnings and error messages

#### Recognize Common Activities and Reduce Click-Depth

<img src="https://user-images.githubusercontent.com/5710611/77146854-62685280-6acf-11ea-98b5-6652f09ac977.png"  width="400px"/>

The pareto principle of user interface design says that 80% of usage is related to 20% of functionality. We have found out that this phenomenon holds true also for the console, where ~70% of users use the console to "check traffic and status". In that sense, we want to ensure that important features are easily accessible. We achieve this through:

- using overview pages in general, as well as for entities, that feature the most important information at a glance
- periodically reevaluating information to be placed in overview sections
- employ overview widgets that show most important info, while referencing advanced info inside respective subviews
