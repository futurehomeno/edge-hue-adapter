import sys

def set_deb_control(version , arch,file_name):
    template = "Package: hue-ad\n"
    template+= "Version: "+version+"\n"
    template+= "Replaces: hue-ad\n"
    template+= "Section: non-free/misc\n"
    template+= "Priority: optional\n"
    template+= "Architecture: "+arch+"\n"
    template+= "Maintainer: Aleksandrs Livincovs <aleksandrs.livincovs@gmail.com>\n"
    template+= "Description: Philips Hue adapter for futurehome system .\n"

    f = open(file_name,"w")
    f.write(template)
    f.close()


def set_version_file(version):
    file_name = "./VERSION"
    f = open(file_name,"w")
    f.write(version)
    f.close()    


if __name__ == "__main__":
   environment = sys.argv[1] 
   version = sys.argv[2]
   arch = sys.argv[3]
   set_deb_control(version,arch,"./package/debian_fh/DEBIAN/control")
   set_deb_control(version,arch,"./package/debian_tp/DEBIAN/control")
   set_version_file(version)