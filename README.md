This is a stateless server which automatically generates api endpoints based on stored functions in a specific directory. The functions should take cli args as payloads or parameters and this should be built automatically. We will also have endpoints generated in a chain from an initial "auth" endpoint which takes a username/password pair. To continue using the system requires leveraging a Rest endpoint. This should be the maximum on the Restful maturity model where the server needs no information. 

This sould be a level 3 AIP as described hre:

Hypermedia as the engine of application state (HATEOAS) is a constraint of the REST software architectural style that distinguishes it from other network architectural styles.[1]

With HATEOAS, a client interacts with a network application whose application servers provide information dynamically through hypermedia. A REST client needs little to no prior knowledge about how to interact with an application or server beyond a generic understanding of hypermedia.

By contrast, clients and servers in Common Object Request Broker Architecture (CORBA) interact through a fixed interface shared through documentation or an interface description language (IDL).

The restrictions imposed by HATEOAS decouple client and server. This enables server functionality to evolve independently.

The term was coined in 2000 by Roy Fielding in his doctoral dissertation.[2]
Example

A user-agent makes an HTTP request to a REST API through an entry point URL. All subsequent requests the user-agent may make are discovered inside the response to each request. The media types used for these representations, and the link relations they may contain, are part of the API. The client transitions through application states by selecting from the links within a representation or by manipulating the representation in other ways afforded by its media type. In this way, RESTful interaction is driven by hypermedia, rather than out-of-band information.[3]

For example, this GET request fetches an account resource, requesting details in a JSON representation:[4]

GET /accounts/12345 HTTP/1.1
Host: bank.example.com

The response is:

HTTP/1.1 200 OK

{
    "account": {
        "account_number": 12345,
        "balance": {
            "currency": "usd",
            "value": 100.00
        },
        "links": {
            "deposits": "/accounts/12345/deposits",
            "withdrawals": "/accounts/12345/withdrawals",
            "transfers": "/accounts/12345/transfers",
            "close-requests": "/accounts/12345/close-requests"
        }
    }
}

The response contains these possible follow-up links: POST a deposit, withdrawal, transfer, or close request (to close the account).

As an example, later, after the account has been overdrawn, there is a different set of available links, because the account is overdrawn.

HTTP/1.1 200 OK

{
    "account": {
        "account_number": 12345,
        "balance": {
            "currency": "usd",
            "value": -25.00
        },
        "links": {
            "deposits": "/accounts/12345/deposits"
        }
    }
}

Now only one link is available: to deposit more money (by POSTing to deposits). In its current state, the other links are not available. Hence the term Engine of Application State. What actions are possible varies as the state of the resource varies.

A client does not need to understand every media type and communication mechanism offered by the server. The ability to understand new media types may be acquired at run-time through "code-on-demand" provided to the client by the server.[2]
Origins

The HATEOAS constraint is an essential part of the "uniform interface" feature of REST, as defined in Roy Fielding's doctoral dissertation.[2] Fielding has further described the concept on his blog.[3]

The purpose of some of the strictness of this and other REST constraints, Fielding explains, is "software design on the scale of decades: every detail is intended to promote software longevity and independent evolution. Many of the constraints are directly opposed to short-term efficiency. Unfortunately, people are fairly good at short-term design, and usually awful at long-term design".[3]
Implementations
Hypertext

    HTML itself is hypermedia, with the <form>...</form> element in control of HTTP requests to links.[3][5] Htmx introduces extensions to HTML to allow elements other than <form>...</form> and <a>...</a> to control requests.

JSON/XML

    HAL, hypermedia built on top of JSON or XML. Defines links, but not actions (HTTP requests).
    JSON-LD, standard for hyperlinks in JSON. Does not address actions.
        Hydra. Builds on top of JSON-LD to add definition of actions.[6]
    Siren, hypermedia built on top of JSON. Defines links and actions.[7]
    Collection+JSON, hypermedia built on top of JSON. Defines links and actions.[8]
    JSON:API, defines links and actions.[9]