Vagrant.configure('2') do |config|
    # grab Fedora 23 image
    config.vm.box = " fedora/23-cloud-base"

    # fix issues with slow dns http://serverfault.com/a/595010
    config.vm.provider :virtualbox do |vb, override|
        vb.customize ["modifyvm", :id, "--natdnshostresolver1", "on"]
        vb.customize ["modifyvm", :id, "--natdnsproxy1", "on"]
        # add more ram, the default isn't enough for the build
        vb.customize ["modifyvm", :id, "--memory", "768"]
    end

    # install Build Dependencies (GOLANG)
    config.vm.provision :shell, :privileged => false, :path => "scripts/vagrant-install-go.sh"

    # Install acbuild
    config.vm.provision :shell, :privileged => false, :path => "scripts/vagrant-install-acbuild.sh"
end
