import os
import hashlib
from typing import List, Optional, Union, Any
from PIL import Image
import natsort

class Robohash:
    """
    Robohash is a quick way of generating unique avatars for a site.
    The original use-case was to create somewhat memorable images to represent a RSA key.
    """

    def __init__(self, string: str, hashcount: int = 11, ignoreext: bool = True):
        """
        Creates our Robohasher
        Takes in the string to make a Robohash out of.
        
        Args:
            string: The input string to hash
            hashcount: Number of hash segments to create
            ignoreext: Whether to ignore file extensions in the string
        """

        # Default to png
        self.format = 'png'

        # Optionally remove an images extension before hashing.
        if ignoreext is True:
            string = self._remove_exts(string)

        string = string.encode('utf-8')

        hash_obj = hashlib.sha512()
        hash_obj.update(string)
        self.hexdigest = hash_obj.hexdigest()
        self.hasharray: List[int] = []
        # Start this at 4, so earlier is reserved
        # 0 = Color
        # 1 = Set
        # 2 = bgset
        # 3 = BG
        self.iter = 4
        self._create_hashes(hashcount)

        self.resourcedir = f"{os.path.dirname(__file__)}/"
        # Get the list of backgrounds and RobotSets
        self.sets = self._listdirs(f"{self.resourcedir}sets")
        self.bgsets = self._listdirs(f"{self.resourcedir}backgrounds")

        # Get the colors in set1
        self.colors = self._listdirs(f"{self.resourcedir}sets/set1")

    def _remove_exts(self, string: str) -> str:
        """
        Sets the string, to create the Robohash
        
        Args:
            string: Input string that may contain an image extension
            
        Returns:
            string with extension removed if present
        """

        # If the user hasn't disabled it, we will detect image extensions, such as .png, .jpg, etc.
        # We'll remove them from the string before hashing.
        # This ensures that /Bear.png and /Bear.bmp will send back the same image, in different formats.

        if string.lower().endswith(('.png','.gif','.jpg','.bmp','.jpeg','.ppm','.datauri')):
            format_str = string[string.rfind('.') + 1:]
            if format_str.lower() == 'jpg':
                format_str = 'jpeg'
            self.format = format_str
            string = string[:string.rfind('.')]
        return string


    def _create_hashes(self, count: int) -> None:
        """
        Breaks up our hash into slots, so we can pull them out later.
        Essentially, it splits our SHA/MD5/etc into X parts.
        
        Args:
            count: Number of segments to split the hash into
        """
        for i in range(count):
             # Get 1/numblocks of the hash
             blocksize = int(len(self.hexdigest) / count)
             currentstart = (1 + i) * blocksize - blocksize
             currentend = (1 + i) * blocksize
             self.hasharray.append(int(self.hexdigest[currentstart:currentend], 16))

        # Workaround for adding more sets in 2019.
        # We run out of blocks, because we use some for each set, whether it's called or not.
        # I can't easily change this without invalidating every hash so far :/
        # This shouldn't reduce the security since it should only draw from one set of these in practice.
        self.hasharray = self.hasharray + self.hasharray

    def _listdirs(self, path: str) -> List[str]:
        """
        Get a list of directories at the given path
        
        Args:
            path: Path to search for directories
            
        Returns:
            List of directory names (not full paths)
        """
        return [d for d in natsort.natsorted(os.listdir(path)) if os.path.isdir(os.path.join(path, d))]

    def _get_list_of_files(self, path: str) -> List[str]:
        """
        Go through each subdirectory of `path`, and choose one file from each to use in our hash.
        Continue to increase self.iter, so we use a different 'slot' of randomness each time.
        
        Args:
            path: Root directory to search for image files
            
        Returns:
            List of chosen file paths, one from each subdirectory
        """
        chosen_files = []

        # Get a list of all subdirectories
        directories = []
        for root, dirs, files in natsort.natsorted(os.walk(path, topdown=False)):
            for name in dirs:
                if not name.startswith('.'):
                    directories.append(os.path.join(root, name))
                    directories = natsort.natsorted(directories)

        # Go through each directory in the list, and choose one file from each.
        # Add this file to our master list of robotparts.
        for directory in directories:
            files_in_dir = []
            for imagefile in natsort.natsorted(os.listdir(directory)):
                files_in_dir.append(os.path.join(directory, imagefile))
                files_in_dir = natsort.natsorted(files_in_dir)

            # Use some of our hash bits to choose which file
            element_in_list = self.hasharray[self.iter] % len(files_in_dir)
            chosen_files.append(files_in_dir[element_in_list])
            self.iter += 1

        return chosen_files

    def assemble(self, 
               roboset: Optional[str] = None, 
               color: Optional[str] = None, 
               format: Optional[str] = None, 
               bgset: Optional[str] = None, 
               sizex: int = 300, 
               sizey: int = 300) -> None:
        """
        Build our Robot!
        Returns the robot image itself in self.img.
        
        Args:
            roboset: Which robot set to use ('set1', 'set2', etc. or 'any')
            color: Color to use (only works with set1)
            format: Image format ('png', 'jpeg', etc.)
            bgset: Background set to use
            sizex: Width of the final image
            sizey: Height of the final image
        """
        # Allow users to manually specify a robot 'set' that they like.
        # Ensure that this is one of the allowed choices, or allow all
        # If they don't set one, take the first entry from sets above.

        if roboset == 'any':
            roboset = self.sets[self.hasharray[1] % len(self.sets)]
        elif roboset in self.sets:
            # No need to reassign roboset to itself
            pass
        else:
            roboset = self.sets[0]

        # Only set1 is setup to be color-selectable. The others don't have enough pieces in various colors.
        # This could/should probably be expanded at some point..
        # Right now, this feature is almost never used. (It was < 44 requests this year, out of 78M reqs)

        if roboset == 'set1':
            if color in self.colors:
                roboset = f"set1/{color}"
            else:
                randomcolor = self.colors[self.hasharray[0] % len(self.colors)]
                roboset = f"set1/{randomcolor}"

        # If they specified a background, ensure it's legal, then give it to them.
        if bgset in self.bgsets:
            # No need to reassign bgset to itself
            pass
        elif bgset == 'any':
            bgset = self.bgsets[self.hasharray[2] % len(self.bgsets)]

        # If we set a format based on extension earlier, use that. Otherwise, PNG.
        if format is None:
            format = self.format

        # Each directory in our set represents one piece of the Robot, such as the eyes, nose, mouth, etc.

        # Each directory is named with two numbers - The number before the # is the sort order.
        # This ensures that they always go in the same order when choosing pieces, regardless of OS.

        # The second number is the order in which to apply the pieces.
        # For instance, the head has to go down BEFORE the eyes, or the eyes would be hidden.

        # First, we'll get a list of parts of our robot.
        roboparts = self._get_list_of_files(f"{self.resourcedir}sets/{roboset}")
        
        # Now that we've sorted them by the first number, we need to sort each sub-category by the second.
        roboparts.sort(key=lambda x: x.split("#")[1])
        
        background = None
        if bgset is not None:
            bglist = []
            backgrounds = natsort.natsorted(os.listdir(f"{self.resourcedir}backgrounds/{bgset}"))
            backgrounds.sort()
            for ls in backgrounds:
                if not ls.startswith("."):
                    bglist.append(f"{self.resourcedir}backgrounds/{bgset}/{ls}")
            background = bglist[self.hasharray[3] % len(bglist)]

        # Paste in each piece of the Robot.
        roboimg = Image.open(roboparts[0])
        roboimg = roboimg.resize((1024, 1024))
        for png in roboparts:
            img = Image.open(png)
            img = img.resize((1024, 1024))
            roboimg.paste(img, (0, 0), img)

        if bgset is not None and background is not None:
            bg = Image.open(background)
            bg = bg.resize((1024, 1024))
            bg.paste(roboimg, (0, 0), roboimg)
            roboimg = bg

        # If we're a BMP, flatten the image.
        if format in ['bmp', 'jpeg']:
            # Flatten bmps
            r, g, b, a = roboimg.split()
            roboimg = Image.merge("RGB", (r, g, b))

        self.img = roboimg.resize((sizex, sizey), Image.LANCZOS)
        self.format = format

