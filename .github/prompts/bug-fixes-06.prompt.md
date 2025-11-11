There are a number of bugs in the current build. Use this prompt to create a phased approach and test driven development to work through this bug. 

## Bug 05

There is some incomplete attributes in our conversion to hcl. It seems some properties on `pingone_davinci_flow` convertions need empty attributes to be maintained. Identify where we are maintaining this in schemas and ensure that when converting to hcl we maintain these empty attributes.

```log
│ 117692:     js_links = [
│ 117693:       {
│ 117694:         label          = "https://ajax.googleapis.com/ajax/libs/jquery/3.6.0/jquery.min.js"
│ 117695:         value          = "https://ajax.googleapis.com/ajax/libs/jquery/3.6.0/jquery.min.js"
│ 117696:       },
│ 117697:       {
│ 117698:         label          = "https://www.google.com/recaptcha/api.js"
│ 117699:         value          = "https://www.google.com/recaptcha/api.js"
│ 117700:       },
│ 117701:       {
│ 117702:         label          = "https://www.google.com/recaptcha/api.js?render=6LfdK6QpAAAAALcGPNmzzyK4Baigr2UWjnL57ZIr"
│ 117703:         value          = "https://www.google.com/recaptcha/api.js?render=6LfdK6QpAAAAALcGPNmzzyK4Baigr2UWjnL57ZIr"
│ 117704:       }
│ 117705:     ]
│ 117706:     log_level                            = 3
│ 117707:     use_custom_css                       = true
│ 117708:     use_custom_script                    = true
│ 117709:   }
│ 
│ Inappropriate value for attribute "settings": attribute "js_links": element 0: attributes "crossorigin",
│ "defer", "integrity", "referrerpolicy", and "type" are required.
```