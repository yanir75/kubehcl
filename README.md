
# kubehcl

## Credit
kubehcl was inspired by two tools helm and terraform/opentofu.  
Parts of the code were copied and modified from the following repositories.   
Links: [helm](https://github.com/helm/helm)  
&nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp;[opentofu](https://github.com/opentofu/opentofu)  
&nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp;[kubectl-validate](https://github.com/kubernetes-sigs/kubectl-validate)  

  
Both licenses were copied from the reflective projects with copyrights within them.  

**This project is not affiliated or endorsed by any of the projects/companies stated above or any other project/company at all.**

## Tool usage
kubehcl renders HCL files and deploys them to kubernetes.  

## Functions
HCL used by this tool contains all the **functions** which exist in opentofu, for example: merge, base64encode etc.

## Blocks
There are 5 kinds of blocks allowed in the configuration:   
**variable** block contains three attributes description, type and default.  
```
description: explaination of the variable usage, optional
```  
```
type: constraint on the variable value, must match the type, optional
```  
```
default: value which will be assigned to the value if not defined elsewhere
```
variables can be used in configuration files such as var.foo

---
**locals** block contains attribute names and their values.  
locals can be used in configuration files such as local.foo

---
**default_annotations** block contains only attributes and string values  
All annotations will be added to the resources in the same level of configuration

---
**resource** block contains the configuration of the kubernetes resource further examples can be seen in [example](/example) folder.  
This block can contain for_each or count and depends_on attributes.
```
for_each contains a map or set of strings which will create a resource for each key
attributes of for each can be accessed by using each.key or each.value accordingly
```
```
count must be a positive number or zero and can be accessed by using count.index
```
```
depends_on list of dependencies can contain only modules or resources
```
---
**module** block contains must have source attribute which is the path to all other configuration files.  
This block can contain for_each or count and depends_on attributes.
```
for_each contains a map or set of strings which will create a module for each key
attributes of for each can be accessed by using each.key or each.value accordingly
```
```
count must be a positive number or zero and can be accessed by using count.index
```
```
depends_on list of dependencies can contain only modules or resources
```

## Vars file
In case you didn't use defaults vars file can be configured, the filename must be kubehcl.tfvars only attributes and values can be assigned to it.  
kubehcl will automatically search this filename and assigne the values to the variables accordingly.
## License Information

This project includes code that is licensed under the **Mozilla Public License 2.0** (MPL-2.0) and the **Apache License 2.0**.

- Parts of this project are licensed under the [Mozilla Public License 2.0](https://www.mozilla.org/en-US/MPL/2.0/).
- Parts of this project are licensed under the [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0).