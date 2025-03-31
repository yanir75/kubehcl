from filelist import file_dict

def line_prepender(filename, line):
    with open(filename, 'r+') as f:
        content = f.read()
        f.seek(0, 0)
        f.write(line.rstrip('\r\n') + '\n' + content)

comment = """/* 
This file was inspired from {repo}
This file has been modified from the original version
Changes made to fit kubehcl purposes
This file retains its' original license
{spdx}
Licesne: {license}
*/
"""

copy_comment = """/* 
{spdx}
This file was copied from {repo} and retains the {license}
*/
"""


repo_dict = {
    "kubectl-validate": ("https://github.com/kubernetes-sigs/kubectl-validate","// SPDX-License-Identifier: Apache-2.0","https://www.apache.org/licenses/LICENSE-2.0"),
    "helm": ("https://github.com/helm/helm","// SPDX-License-Identifier: Apache-2.0","https://www.apache.org/licenses/LICENSE-2.0"),
    "opentofu": ("https://github.com/opentofu/opentofu","// SPDX-License-Identifier: MPL-2.0","https://www.mozilla.org/en-US/MPL/2.0/"),
}
for key,value in file_dict.items():
    repo_name = value[0]
    modified = value[1]
    if repo_name !="mine":
        name,spdx,url = repo_dict[repo_name]
        if modified:
            line_prepender(key,comment.format(repo = name, spdx = spdx, license = url)+"\n")
        else:
            line_prepender(key,copy_comment.format(repo = name, spdx = spdx, license = url)+"\n")
