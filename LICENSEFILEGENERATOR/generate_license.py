from filelist import file_dict

def line_prepender(filename, line):
    with open(filename, 'r+') as f:
        content = f.read()
        f.seek(0, 0)
        f.write(line.rstrip('\r\n') + '\n' + content)

comment = """/* This file was inspired from {repo}
This file has been modified from the original version
Changes made to fit kubehcl purposes
*/"""

for key,value in file_dict:
    repoName = value[0]
    modified = value[1]
    if modified and repoName !="mine":
        line_prepender(key,comment.format(repo = repoName.capitalize()))